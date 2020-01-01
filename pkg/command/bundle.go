package command

import (
	"embly/pkg/build"

	flag "github.com/spf13/pflag"
)

type bundleCommand struct {
	flagSet            *flag.FlagSet
	includeObjectFiles *bool
}

func (f *bundleCommand) flags() *flag.FlagSet {
	f.flagSet = &flag.FlagSet{}
	f.includeObjectFiles = f.flagSet.BoolP("include-objects", "i", false, "Include locally built object files in the archive")
	return f.flagSet
}

func (f *bundleCommand) help() string {
	return `
Usage: embly bundle

	Create a bundled project file`
}

func (f *bundleCommand) run(args []string) (err error) {
	f.flags()
	_ = f.flagSet.Parse(args)
	builder, err := build.NewBuilder("", UI)
	if err != nil {
		return
	}
	if len(builder.Config.Gateways) == 0 {
		UI.Info("No gateways, nothing to run")
		return nil
	}
	if err = builder.CompileFunctionsToWasm(); err != nil {
		return
	}
	if *f.includeObjectFiles {
		if err = builder.CompileWasmToObject(); err != nil {
			return
		}
	}

	location := "out.tar.gz"
	if err = builder.Bundle(location, *f.includeObjectFiles); err != nil {
		return
	}
	UI.Info("Wrote project output to " + location)
	return nil
}

func (f *bundleCommand) synopsis() string {
	return "Create a bundled project file"
}
