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

type Master struct {
	functions map[string]string
	registry  map[uint64]funcOrGateway
}

type funcOrGateway interface {
	SendMsg(comms_proto.Message)
}

func NewMaster() *Master {
	return &Master{
		registry:  make(map[uint64]funcOrGateway),
		functions: make(map[string]string),
	}
}

// SockAddr is the location of the embly unix socket
const SockAddr = "/tmp/embly.sock"

func prepareMsg(msg comms_proto.Message) (b []byte, err error) {
	b, err = proto.Marshal(&msg)
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(b)))
	if err != nil {
		return
	}
	b = append(size, b...)
	return
}

type consumer struct {
	source io.Reader
}

func (p *consumer) nextMessage() (msg comms_proto.Message, err error) {
	sizeBytes := make([]byte, 4)
	_, err = p.source.Read(sizeBytes)
	if err != nil {
		return
	}
	size := int(binary.LittleEndian.Uint32(sizeBytes))
	read := 0
	msgBytes := make([]byte, size)
	for {
		var ln int
		if ln, err = p.source.Read(msgBytes[read:]); err != nil {
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

type Function struct {
	addr     uint64
	parent   uint64
	cmd      *exec.Cmd
	conn     net.Conn
	connWait sync.WaitGroup
}

func (fn *Function) RegisterConn(conn net.Conn) {
	fn.conn = conn
	fn.connWait.Done()
}

func (fn *Function) HasConnOrWait() {
	fn.connWait.Wait()
}

func (fn *Function) SendMsg(msg comms_proto.Message) {
	fn.HasConnOrWait()
	b, err := prepareMsg(msg)
	if err != nil {
		log.Println(err)
	}
	// net.Conn is thread safe
	// TODO: ensure all was written
	fn.conn.Write(b)
}

func (gat *Gateway) SendMsg(msg comms_proto.Message) {
	if msg.Exit != 0 {
		log.Fatal("unimplemented")
	}
	if msg.Spawn != "" {
		log.Fatal("unimplemented")
	}
	gat.buf.Write(msg.Data)
	gat.readCond.Broadcast()
}

type Gateway struct {
	ID       uint64
	buf      bytes.Buffer
	readCond *sync.Cond
	child    uint64
	master   *Master
	msgChan  chan comms_proto.Message
}

func (m *Master) NewGateway() *Gateway {
	id := rand.Uint64()
	mu := sync.Mutex{}
	gat := &Gateway{
		ID:       id,
		msgChan:  make(chan comms_proto.Message, 2),
		master:   m,
		readCond: sync.NewCond(&mu),
	}
	m.registry[id] = gat
	return gat
}

func (gat *Gateway) AttachFn(fn *Function) {
	gat.child = fn.addr
}

func (gat *Gateway) Wait() {
	gat.readCond.L.Lock()
	if gat.buf.Len() == 0 {
		gat.readCond.Wait()
	}
	gat.readCond.L.Unlock()
}

func (get *Gateway) Bytes() (b []byte) {
	b = get.buf.Bytes()
	get.buf.Reset()
	return b
}
func (gat *Gateway) Read(b []byte) (ln int, err error) {
	gat.Wait()
	return gat.buf.Read(b)
}

func (gat *Gateway) Write(b []byte) (ln int, err error) {
	log.Println("Write from ", gat.ID, "to", gat.child)
	fn := gat.master.registry[gat.child]
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

func (fn *Function) Start() (err error) {
	return fn.cmd.Start()
}

func (m *Master) RegisterFunctionName(name, location string) {
	m.functions[name] = location
}

func (m *Master) SpawnFunction(name string, parent uint64, addr uint64) {
	fn := m.NewFunction(m.functions[name], parent, &addr)
	fn.Start()
}

func (m *Master) NewFunction(location string, parent uint64, addr *uint64) (fn *Function) {
	if addr == nil {
		v := rand.Uint64()
		addr = &v
	}
	fn = &Function{addr: *addr}
	fn.connWait.Add(1)
	cmd := exec.Command("embly-wrapper-rs")
	log.Println("NewFunction location", location)
	cmd.Stdout = textio.NewPrefixWriter(os.Stdout, "embly stdout: ")
	cmd.Stderr = textio.NewPrefixWriter(os.Stderr, "embly stderr: ")
	cmd.Env = envVars(map[string]string{
		"EMBLY_ADDR":     fmt.Sprintf("%d", fn.addr),
		"EMBLY_MODULE":   location,
		"RUST_BACKTRACE": "1",
	})
	fn.cmd = cmd
	fn.parent = parent
	m.registry[fn.addr] = fn
	return
}

// Start starts listening on ths unix socket and will let fns communicate
func (m *Master) Start() {
	m.unixListen(func(conn net.Conn) {
		addrBytes := make([]byte, 8)
		conn.Read(addrBytes)
		addr := binary.LittleEndian.Uint64(addrBytes)
		fnOrG := m.registry[addr]
		fn := fnOrG.(*Function)

		msg := comms_proto.Message{
			YourAddress:   addr,
			ParentAddress: fn.parent,
		}

		b, err := prepareMsg(msg)
		if err != nil {
			log.Println(err)
		}
		conn.Write(b)
		fn.RegisterConn(conn)
		c := consumer{source: conn}
		for {
			msg, err := c.nextMessage()
			if err != nil {
				log.Println(err)
				if err == io.EOF {
					return
				}
				continue
			}

			log.Printf("got message %#v\n", msg)
			if msg.Spawn != "" {
				m.SpawnFunction(msg.Spawn, msg.From, msg.SpawnAddress)
				continue
			}
			// TODO: security: allows one to communicate with any function
			recFn := m.registry[msg.To]
			if recFn == nil {
				log.Fatal("fn not found for id ", msg.To)
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
