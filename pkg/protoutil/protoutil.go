package protoutil

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/gogo/protobuf/proto"
)

// WriteMessage writes a proto struct to an io.Writer
func WriteMessage(consumer io.Writer, msg proto.Message) (err error) {
	b, err := proto.Marshal(msg)
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(b)))
	if err != nil {
		return
	}
	b = append(size, b...)
	ln, err := consumer.Write(b)
	if ln != len(b) {
		fmt.Println(consumer, "didn't write everything!")
	}
	return

}

// NextMessage grabs the next message from a reader
func NextMessage(consumer io.Reader, msg proto.Message) (err error) {
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
	return proto.Unmarshal(msgBytes, msg)
}

//
