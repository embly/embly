package command

import (
	"embly/pkg/config"
	"embly/pkg/nixbuild"
	"fmt"

	flag "github.com/spf13/pflag"
)

func createBuilder(path string) (builder *nixbuild.Builder, cfg *config.Config, err error) {
	if builder, err = nixbuild.NewClientBuilder(UI); err != nil {
		return
	}
	cfg, err = config.New(path)
	if err != nil {
		return
	}
	builder.SetProject(cfg)
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

func (f *buildCommand) run(args []string) (err error) {
	builder, cfg, err := createBuilder("")
	if err != nil {
		return
	}
	for _, fn := range cfg.Functions {
		var result string
		result, err = builder.Build(fn.Name)
		if err != nil {
			return
		}
		fmt.Println(result)
	}
	return err
}

func (f *buildCommand) synopsis() string {
	return "Build an embly project"
}
