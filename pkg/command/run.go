package command

import (
	"embly/pkg/build"
	"embly/pkg/core"
	"os"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

type runCommand struct {
	flagSet *flag.FlagSet
	host    *string
}

func (f *runCommand) flags() *flag.FlagSet {
	f.flagSet = &flag.FlagSet{}
	f.host = f.flagSet.String("host", "", "set the host to broadcast on")
	return f.flagSet
}
func (f *runCommand) synopsis() string {
	return "Run a local embly project"
}

func (f *runCommand) help() string {
	return `
Usage: embly run [options] (<location>)

<location> can be a local directory, git repo, or bundled embly project eg:

embly run ./
embly run archive.tar
embly run github.com/embly/app/subproject

Run a local embly project. Running without a <location> will default to
the current working directory. If there isn't an embly.hcl in the current
directly embly will crawl all parent directories to see if one exists.`
}

func (f *runCommand) run(args []string) (err error) {
	var location string
	if len(args) == 1 {
		location = args[0]
	} else if len(args) > 1 {
		return errors.New("error: embly run takes only one positional argument")
	}
	var isFile bool
	if location != "" {
		fi, err := os.Stat(location)
		if err != nil {
			return errors.Errorf(`location "%s" doesn't exist`, location)
		}
		isFile = !fi.IsDir()
	}

	var builder *build.Builder

	if isFile {
		builder, err = build.NewBuilderFromArchive(location, UI)
		if err != nil {
			return
		}

		if err = builder.CompileWasmToObject(isFile); err != nil {
			return
		}
	} else {
		// TODO
		// builder, err = runBuild(location)
		// if err != nil {
		// 	return
		// }
	}

	if len(builder.Config.Gateways) == 0 {
		UI.Info("No gateways, nothing to run")
		return nil
	}
	if err := core.Start(builder, UI, core.StartConfig{
		Watch: false,
		Host:  *f.host,
	}); err != nil {
		return err
	}
	return nil
}
