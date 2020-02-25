package main

import (
	"embly/pkg/nixbuild"
	"log"
)

func main() {
	builder, err := nixbuild.NewBuildServer()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("starting server")
	if err := builder.StartServer(); err != nil {
		log.Fatal(err)
	}
}
