package comms

import (
	"bytes"
	comms_proto "embly/pkg/comms/proto"
	"testing"
)

func TestFoo(t *testing.T) {
	msg := comms_proto.Message{
		Exit:  1,
		Spawn: "hello",
	}

	b, err := prepareMsg(msg)
	if err != nil {
		t.Error(err)
	}
	c := consumer{
		source: bytes.NewBuffer(b),
	}
	msgBack, err := c.nextMessage()
	if err != nil {
		t.Error(err)
	}
	if msgBack.Exit != 1 {
		t.Error("exit code should be 1")
	}
}
