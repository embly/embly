package main

import (
	"embly/pkg/nixbuild"
	"log"
)

func main() {
	builder, err := nixbuild.NewBuilder(nixbuild.BuildConfig{})
	if err != nil {
		log.Fatal(err)
	}

	if err := builder.StartServer(); err != nil {
		log.Fatal(err)
	}
}
