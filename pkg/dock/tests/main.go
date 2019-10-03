package main

import (
	"embly/pkg/dock"
	"fmt"
	"log"
)

func main() {
	// v, err := dock.StartVinyl("foo")
	// check(err)
	// _ = v

	b, err := dock.DescriptorForFile("../../core/proto/comms.proto")
	check(err)
	fmt.Printf("%#v\n", b)

	//
	// fmt.Println(string(b))
	// c, err := dock.NewClient()
	// check(err)
	// // err = c.PullImage("python:3-slim")
	// // check(err)

	// cont := c.NewContainer("dock-test", "python:3-slim")
	// cont.Cmd = []string{"sleep", "10000"}
	// cont.WorkingDir = "/opt"
	// check(cont.Stop())
	// check(cont.Remove())
	// check(cont.Create())
	// check(cont.Start())

	// // check(cont.Exec("mkdir -p /var/embly && cd /opt && mkdir foo && touch bar && touch foo/bar"))

	// // check(c.Copy("dock-test:/", "."))
	// // check(c.Copy(".", "dock-test:/var/embly"))
	// // check(c.Copy("./main.go", "dock-test:/var/embly/foo"))
	// // check(c.Copy(".", "dock-test:/var/embly/foo"))
	// check(c.Copy("./main.go", "dock-test:/var/main.go"))

	// check(cont.Exec("ls -lah /var/embly/"))
}

func check(err error) {
	if err != nil {
		fmt.Printf("%+v\n", err)
		log.Fatal(err)
	}
}
