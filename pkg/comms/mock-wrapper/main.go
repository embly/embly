package main

import (
	"embly/pkg/comms"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {
	// the uint64 string value of the addr of this function
	emblyAddr := os.Getenv("EMBLY_ADDR")
	emblySocket := os.Getenv("EMBLY_SOCKET")

	log.Println("Started")
	log.Println("EMBLY_ADDR", emblyAddr)
	log.Println("EMBLY_SOCKET", emblySocket)
	if emblyAddr == "" {
		panic("need addr")
	}
	if emblySocket == "" {
		panic("need socket")
	}
	conn, err := net.Dial("unix", emblySocket)
	if err != nil {
		panic(err)
	}
	addr, err := strconv.ParseUint(emblyAddr, 10, 64)
	if err != nil {
		panic(err)
	}
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, addr)
	_, err = conn.Write(b)
	if err != nil {
		panic(err)
	}

	// startup message
	msg, err := comms.NextMessage(conn)
	if err != nil {
		panic(err)
	}
	if msg.ParentAddress == 0 || msg.YourAddress == 0 {
		panic(msg)
	}

	for {
		msg2, err := comms.NextMessage(conn)
		fmt.Println("mock-wrapper got msg", msg)
		if err != nil {
			panic(err)
		}
		from := msg2.From
		to := msg2.To
		msg2.From = to
		msg2.To = from
		if err := comms.WriteMessage(conn, msg2); err != nil {
			panic(err)
		}
	}
}
