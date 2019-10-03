package core

import (
	"bytes"
	comms_proto "embly/pkg/core/proto"
	"fmt"
	"os/exec"
	"testing"
)

func init() {
	EmblyWrapperExecutable = "mock-wrapper"
	cmd := exec.Command("bash", "-c", "pwd && cd ./mock-wrapper && go install")
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		panic(err)
	}
}

func TestBasicMaster(t *testing.T) {
	m := NewMaster()
	go m.Start()

	gat := m.NewGateway()

	m.functions["foo"] = ""
	fn, err := m.NewFunction("foo", gat.ID, nil, nil)
	if err != nil {
		t.Error(err)
	}
	gat.AttachFn(fn)
	if err := fn.Start(); err != nil {
		t.Error(err)
	}

	sending := []byte("it's lunchtime")
	gat.Write(sending)

	buf := make([]byte, len(sending))
	_, err = gat.Read(buf)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(buf, sending) {
		t.Error("bytes should be equal")
	}

	m.StopFunction(fn)
	if _, ok := m.registry.Load(fn.addr); ok {
		t.Error("function should no longer be around")
	}

	m.RemoveGateway(gat)
	if _, ok := m.registry.Load(gat.ID); ok {
		t.Error("function should no longer be around")
	}

}

func TestFoo(t *testing.T) {
	msg := comms_proto.Message{
		Exit:  1,
		Spawn: "hello",
	}
	buf := &bytes.Buffer{}
	if err := WriteMessage(buf, msg); err != nil {
		t.Error(err)
	}
	msgBack, err := NextMessage(buf)
	if err != nil {
		t.Error(err)
	}
	if msgBack.Exit != 1 {
		t.Error("exit code should be 1")
	}
}
