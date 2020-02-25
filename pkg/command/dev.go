package command

import (
	"embly/pkg/build"
	"embly/pkg/core"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

type devCommand struct {
	flagSet   *flag.FlagSet
	dontWatch *bool
}

func (f *devCommand) flags() *flag.FlagSet {
	f.flagSet = &flag.FlagSet{}
	f.dontWatch = f.flagSet.BoolP("dont-watch", "d", false, "Disable watching for changes on local files and rebuilding")
	return f.flagSet
}
func (f *devCommand) synopsis() string {
	return "Develop a local embly project"
}

func (f *devCommand) help() string {
	return `
Usage: embly dev [options] (<location>)

<location> is an optional local directory

embly dev ./

Run a local embly project for development. This mode is similar to "embly run"
but it rebuilds functions when their source changes and will enable other
development features.

Running without a <location> will default to the current working directory. If
there isn't an embly.hcl in the current directly embly will crawl all parent
directories to see if one exists.`
}

func (f *devCommand) run(args []string) (err error) {
	var location string
	if len(args) == 1 {
		location = args[0]
	} else if len(args) > 1 {
		return errors.New("error: embly dev takes only one positional argument")
	}

	var builder *build.Builder

	_ = location

	if len(builder.Config.Gateways) == 0 {
		UI.Info("No gateways, nothing to run")
		return nil
	}
	if err := core.Start(builder, UI, core.StartConfig{
		Watch: !*f.dontWatch,
		Dev:   true,
	}); err != nil {
		return err
	}
	return nil
}
