package comms

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"sync"

	comms_proto "embly/pkg/comms/proto"

	"github.com/gogo/protobuf/proto"
	"github.com/segmentio/textio"
)

// Master is the central coordinator for embly functions
type Master struct {
	mutex     sync.Mutex
	registry  sync.Map
	functions map[string]string
}

type funcOrGateway interface {
	SendMsg(comms_proto.Message)
}

// NewMaster creates a new master
func NewMaster() *Master {
	return &Master{
		registry:  sync.Map{},
		functions: make(map[string]string),
	}
}

// SockAddr is the location of the embly unix socket
var SockAddr = "/tmp/embly.sock"

// EmblyWrapperExecutable is the executable we'll run
var EmblyWrapperExecutable = "embly-wrapper"

// WriteMessage writes a proto Message to an io.Writer
func WriteMessage(consumer io.Writer, msg comms_proto.Message) (err error) {
	b, err := proto.Marshal(&msg)
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(b)))
	if err != nil {
		return
	}
	b = append(size, b...)
	ln, err := consumer.Write(b)
	if ln != len(b) {
		panic("didn't write everything!")
	}
	return
}

// NextMessage grabs the next message from a reader
func NextMessage(consumer io.Reader) (msg comms_proto.Message, err error) {
	sizeBytes := make([]byte, 4)
	_, err = consumer.Read(sizeBytes)
	if err != nil {
		return
	}
	size := int(binary.LittleEndian.Uint32(sizeBytes))
	read := 0
	msgBytes := make([]byte, size)
	for {
		var ln int
		if ln, err = consumer.Read(msgBytes[read:]); err != nil {
			return
		}
		read += ln
		if read == size {
			break
		}
		if read > size {
			log.Fatal("not ok")
		}
	}
	err = proto.Unmarshal(msgBytes, &msg)
	return
}

// Function handles the state and connection for an embly function
type Function struct {
	addr     uint64
	parent   uint64
	cmd      *exec.Cmd
	conn     net.Conn
	exited   int32
	connWait sync.WaitGroup
}

// RegisterConn registers a unix socket connection for this conn
func (fn *Function) RegisterConn(conn net.Conn) {
	fn.conn = conn
	fn.connWait.Done()
}

// HasConnOrWait will wait if there isn't a connection associated with this function yet
func (fn *Function) HasConnOrWait() {
	fn.connWait.Wait()
}

// SendMsg sends a protobuf Message to this function
func (fn *Function) SendMsg(msg comms_proto.Message) {
	fn.HasConnOrWait()
	if err := WriteMessage(fn.conn, msg); err != nil {
		log.Println(err)
	}
}

// A Gateway is a way for a function to communicate with the outside world
type Gateway struct {
	ID          uint64
	bufMutex    sync.Mutex
	buf         bytes.Buffer
	readCond    *sync.Cond
	child       uint64
	master      *Master
	childExited int32
	msgChan     chan comms_proto.Message
}

// NewGateway creates a new gateway
func (m *Master) NewGateway() *Gateway {
	id := rand.Uint64()
	mu := sync.Mutex{}
	gat := &Gateway{
		ID:          id,
		msgChan:     make(chan comms_proto.Message, 2),
		master:      m,
		childExited: -1, // running
		readCond:    sync.NewCond(&mu),
	}
	m.addFuncOrGateway(id, gat)
	return gat
}

// RemoveGateway removes a gateway
func (m *Master) RemoveGateway(gat *Gateway) {
	m.delFuncOrGateway(gat.ID)
}

// AttachFn attaches a function to this gateway
func (gat *Gateway) AttachFn(fn *Function) {
	gat.child = fn.addr
}

// SendMsg sends a protobuf Message to this gateway
func (gat *Gateway) SendMsg(msg comms_proto.Message) {
	if msg.Exiting {
		gat.childExited = msg.Exit
		gat.readCond.Broadcast()
	} else if msg.Spawn != "" {
		log.Fatal("unimplemented")
	} else {
		gat.bufMutex.Lock()
		gat.buf.Write(msg.Data)
		gat.bufMutex.Unlock()
		gat.readCond.Broadcast()
	}
}

// Wait waits for bytes to be available to be read from the gateway
func (gat *Gateway) Wait() {
	gat.readCond.L.Lock()
	if gat.buf.Len() == 0 && gat.childExited == -1 {
		fmt.Println("waiting!")
		gat.readCond.Wait()
		fmt.Println("released!")
	}
	gat.readCond.L.Unlock()
}

// Bytes dumps all available bytes from the gateway
func (gat *Gateway) Bytes() (b []byte) {
	gat.bufMutex.Lock()
	b = gat.buf.Bytes()
	gat.bufMutex.Unlock()
	gat.buf.Reset()
	return b
}

func (gat *Gateway) Read(b []byte) (ln int, err error) {
	gat.Wait()
	gat.bufMutex.Lock()
	defer gat.bufMutex.Unlock()
	return gat.buf.Read(b)
}

func (gat *Gateway) Write(b []byte) (ln int, err error) {
	log.Println("Write from ", gat.ID, "to", gat.child)
	fn := gat.master.getFuncOrGateway(gat.child)
	msg := comms_proto.Message{
		To:   gat.child,
		From: gat.ID,
		Data: b,
	}
	fn.SendMsg(msg)
	ln = len(b)
	return
}

func envVars(values map[string]string) (out []string) {
	for k, v := range values {
		out = append(out, k+"="+v)
	}
	return
}

// Start starts a functions process
func (fn *Function) Start() (err error) {
	return fn.cmd.Start()
}

// Stop a functions process
func (fn *Function) Stop() {
	if fn.cmd.Process != nil {
		err := fn.cmd.Process.Kill()
		if err != nil {
			log.Println("error killing func", err)
		}
	}
}

// RegisterFunctionName takes an object file location and a function name for future reference
func (m *Master) RegisterFunctionName(name, location string) {
	m.functions[name] = location
}

// SpawnFunction creates a starts a function with a provided address
func (m *Master) SpawnFunction(name string, parent uint64, addr uint64) error {
	fn := m.NewFunction(m.functions[name], parent, &addr)
	return fn.Start()
}

// StopFunction stops a function and removes it from the registry
func (m *Master) StopFunction(fn *Function) {
	fn.Stop()
	// TODO: not threadsafe?
	m.delFuncOrGateway(fn.addr)
}

func (m *Master) getFuncOrGateway(addr uint64) funcOrGateway {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	fog, _ := m.registry.Load(addr)
	return fog.(funcOrGateway)
}
func (m *Master) addFuncOrGateway(addr uint64, fog funcOrGateway) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.registry.Store(addr, fog)
}

func (m *Master) delFuncOrGateway(addr uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.registry.Delete(addr)
}

// NewFunction creates and initializes a new function, it doesn't start until function.Start is run
func (m *Master) NewFunction(location string, parent uint64, addr *uint64) (fn *Function) {
	if addr == nil {
		v := rand.Uint64()
		addr = &v
	}
	fn = &Function{addr: *addr}
	fn.connWait.Add(1)
	cmd := exec.Command(EmblyWrapperExecutable)
	log.Println("NewFunction location", location)
	cmd.Stdout = textio.NewPrefixWriter(os.Stdout, "embly stdout: ")
	cmd.Stderr = textio.NewPrefixWriter(os.Stderr, "embly stderr: ")
	cmd.Env = envVars(map[string]string{
		"EMBLY_ADDR":     fmt.Sprintf("%d", fn.addr),
		"EMBLY_SOCKET":   SockAddr,
		"EMBLY_MODULE":   location,
		"RUST_BACKTRACE": "1",
		"RUST_LOG":       "embly_wrapper",
	})
	fn.cmd = cmd
	fn.parent = parent
	m.addFuncOrGateway(fn.addr, fn)
	return
}

func (m *Master) functionStartProcess(conn net.Conn) (err error) {
	addrBytes := make([]byte, 8)
	ln, err := conn.Read(addrBytes)
	if err != nil {
		log.Fatal(err)
	}
	if ln != 8 {
		log.Fatalf("incorrect read length %d", ln)
	}
	addr := binary.LittleEndian.Uint64(addrBytes)
	fnOrG := m.getFuncOrGateway(addr)
	// we don't get unix messages from gateways
	fn := fnOrG.(*Function)

	msg := comms_proto.Message{
		YourAddress:   addr,
		ParentAddress: fn.parent,
	}

	if err := WriteMessage(conn, msg); err != nil {
		log.Println(err)
	}
	fn.RegisterConn(conn)
	return
}

// Start starts listening on ths unix socket and will let fns communicate
func (m *Master) Start() error {
	return m.unixListen(func(conn net.Conn) {
		if err := m.functionStartProcess(conn); err != nil {
			log.Fatal(err)
		}

		for {
			msg, err := NextMessage(conn)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println(err)
				continue
			}

			log.Printf("got message %#v %s\n", msg, string(msg.Data))
			if msg.Spawn != "" {
				if err := m.SpawnFunction(msg.Spawn, msg.From, msg.SpawnAddress); err != nil {
					log.Fatal(err)
				}
				continue
			}
			// TODO: security: allows one to communicate with any function
			recFn := m.getFuncOrGateway(msg.To)
			if recFn == nil {
				log.Fatal("fn not found for id ", msg.To)
				continue
			}

			if msg.Exiting {
				log.Println("Function exiting with code", msg.Exit)
			}

			recFn.SendMsg(msg)
		}
	})
}

func (m *Master) unixListen(handler func(net.Conn)) (err error) {
	if err := os.RemoveAll(SockAddr); err != nil {
		return err
	}
	l, err := net.Listen("unix", SockAddr)
	if err != nil {
		return err
	}
	log.Println("listening on " + SockAddr)
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handler(conn)
	}
}
