package comms

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestBasic(t *testing.T) {
	var ln int
	var err error
	c := NewCommGroup()

	readBuf := make([]byte, 10)
	ln, err = c.Read(1, readBuf)
	if err != io.EOF {
		t.Error(errors.New("should have EOF on first read"))
	}

	initData := []byte("hello")
	ln, err = c.MasterWrite(1, initData)
	if ln != len(initData) {
		t.Error(errors.New("len for initData write is wrong"))
	}

	b := make([]byte, 10)
	ln, err = c.Read(1, b)
	if err != nil {
		t.Error(err)
	}
	if ln != len(initData) {
		t.Error("wrong ln for read")
	}

	ln, err = c.Read(1, b)
	if err != io.EOF {
		t.Error(errors.New("expected EOF"))
	}
	if ln != 0 {
		t.Error(errors.New("expected 0 read"))
	}

	var oknow bool
	go func() {
		c.Read(1, b)
		if oknow != true {
			t.Fatal("should block")
		}
	}()

	// let thread preemption kick in so that the goroutine is run
	time.Sleep(time.Millisecond)
	oknow = true
	c.MasterWrite(1, initData)

}

func TestReadAll(t *testing.T) {
	cg := NewCommGroup()
	var ln int
	var err error

	initData := []byte("hello hello hello hello hello hello hello hello hello hello hello hello hello hello")
	ln, err = cg.MasterWrite(1, initData)
	if ln != len(initData) {
		t.Error(errors.New("len for initData write is wrong"))
	}
	if err != nil {
		t.Error(err)
	}

	c, err := cg.GetComm(1)
	if err != nil {
		t.Error(err)
	}
	b, err := ioutil.ReadAll(c)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(b, initData) {
		t.Error("bytes should be equal")
	}
}
