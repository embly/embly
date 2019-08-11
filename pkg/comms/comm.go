package comms

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/pkg/errors"
)

// CommGroup is for function communication
//
// Our comm is for our function and for our orchestrator. We'll call them "master" and "function".
// The function should be able to read bytes sent to the function. It should be able to call read
// and it will recieve EOF if there is nothing there or if it has read all bytes. The next call will
// block and wait for bytes.
//
// The function can write whenever it wants, write operations always complete immediately and never
// block.
//
// The function can also spawn a function and read/write to the spawned function. Read and write rules
// for spawned functions are the same as above.
//
// The master similar rules. It can write and those writes will be buffered. It can read to EOF or
// block for future bytes.
//
// A function or a master should also be able to block until there is activity from any source.
type CommGroup interface {
	GetComm(int) (*Comm, error)
	Read(int, []byte) (int, error)
	Write(int, []byte) (int, error)
	MasterRead(int, []byte) (int, error)
	MasterWrite(int, []byte) (int, error)
	MasterWaitOnWrite() int // TODO
	NewComm() int
}

type defaultCommGroup struct {
	CommGroup
	comms map[int]*Comm
}

type commReadWriter struct {
	buf  *bytes.Buffer
	iofd bool
	cond *sync.Cond
}

func (c *commReadWriter) Read(b []byte) (ln int, err error) {
	c.cond.L.Lock()
	if c.iofd {
		fmt.Println("waiting")
		c.cond.Wait()
	}
	ln, err = c.buf.Read(b)
	if err == io.EOF {
		c.iofd = true
	}
	c.cond.L.Unlock()
	return
}

func (c *commReadWriter) Write(b []byte) (ln int, err error) {
	c.cond.L.Lock()
	ln, err = c.buf.Write(b)
	if c.iofd {
		c.iofd = false
		fmt.Println("broadcasting")
		c.cond.Broadcast()
	}
	c.cond.L.Unlock()
	return
}

func newCommReadWriter() *commReadWriter {
	m := sync.Mutex{}
	c := sync.NewCond(&m)
	return &commReadWriter{
		buf:  bytes.NewBuffer([]byte{}),
		iofd: false,
		cond: c,
	}
}

// Comm is read and write buffers for a single resource
type Comm struct {
	reader *commReadWriter
	writer *commReadWriter
}

// NewComm creates a new comm
func NewComm() *Comm {
	return &Comm{
		reader: newCommReadWriter(),
		writer: newCommReadWriter(),
	}
}

func (c *Comm) Read(b []byte) (ln int, err error) {
	return c.reader.Read(b)
}

func (c *Comm) Write(b []byte) (ln int, err error) {
	return c.writer.Write(b)
}

// MasterWrite allows the master to write to the reader
func (c *Comm) MasterWrite(b []byte) (ln int, err error) {
	return c.reader.Write(b)
}

// MasterRead allows the master to read from the writer
func (c *Comm) MasterRead(b []byte) (ln int, err error) {
	return c.writer.Read(b)
}

func (c *Comm) MasterReader() io.Reader {
	return &masterReader{
		Comm: c,
	}
}

type masterReader struct {
	*Comm
}

func (c *masterReader) Read(b []byte) (ln int, err error) {
	return c.Comm.MasterRead(b)
}

func (c *masterReader) Write(b []byte) (ln int, err error) {
	return c.Comm.MasterWrite(b)
}

// NewCommGroup creates a new comm
func NewCommGroup() CommGroup {
	return &defaultCommGroup{
		comms: map[int]*Comm{
			1: NewComm(), // what is an array if not a map of ints to values...
		},
	}
}

// ErrNoComm no comm exists with this id
var ErrNoComm = errors.New("no comm exists")

func (cg *defaultCommGroup) GetComm(id int) (c *Comm, err error) {
	var ok bool
	if c, ok = cg.comms[id]; ok == false {
		return nil, ErrNoComm
	}
	return c, nil
}

func (cg *defaultCommGroup) Read(id int, b []byte) (ln int, err error) {
	var c *Comm
	var ok bool
	if c, ok = cg.comms[id]; ok == false {
		return 0, ErrNoComm
	}
	return c.reader.Read(b)
}

func (cg *defaultCommGroup) Write(id int, b []byte) (ln int, err error) {
	var c *Comm
	var ok bool
	if c, ok = cg.comms[id]; ok == false {
		c = NewComm()
		cg.comms[id] = c
	}
	return c.writer.Write(b)
}

func (cg *defaultCommGroup) MasterWrite(id int, b []byte) (ln int, err error) {
	var c *Comm
	var ok bool
	if c, ok = cg.comms[id]; ok == false {
		c = NewComm()
		cg.comms[id] = c
	}
	return c.reader.Write(b)
}

func (cg *defaultCommGroup) MasterRead(id int, b []byte) (ln int, err error) {
	var c *Comm
	var ok bool
	if c, ok = cg.comms[id]; ok == false {
		return 0, ErrNoComm
	}
	return c.writer.Read(b)
}

func (cg *defaultCommGroup) NewComm() (ln int) {
	id := len(cg.comms) + 1
	cg.comms[id] = NewComm()
	return id
}
