package command

import (
	"embly/pkg/build"

	flag "github.com/spf13/pflag"
)

func runBuild(path string) (builder *build.Builder, err error) {
	if builder, err = build.NewBuilder(path, UI); err != nil {
		return
	}
	if err = builder.CompileFunctions(); err != nil {
		return
	}
	return
}

type buildCommand struct{}

func (f *buildCommand) flags() *flag.FlagSet {
	return nil
}

func (f *buildCommand) help() string {
	return `
Usage: embly build (<function-name>)...

    Build a local embly project. 
	`
}

func (f *buildCommand) run(args []string) error {
	_, err := runBuild("")
	return err
}

func (f *buildCommand) synopsis() string {
	return "Build an embly project"
}
