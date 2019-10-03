package command

import (
	"os"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
)

type meta struct {
	ui     cli.Ui
	logger hclog.Logger
}

func newMeta(ui cli.Ui, logger hclog.Logger) *meta {
	return &meta{
		ui:     ui,
		logger: logger,
	}
}

// Run the embly command line
func Run() {
	ui := &cli.ColoredUi{
		Ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
		InfoColor: cli.UiColor{
			Bold: true,
		},
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "embly",
		Output: &cli.UiWriter{Ui: ui},
	})

	c := cli.NewCLI("embly", "0.0.1")
	c.Args = os.Args[1:]
	meta := newMeta(ui, logger)

	c.Commands = map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &buildCommand{
				meta: meta,
			}, nil
		},
		"run": func() (cli.Command, error) {
			return &runCommand{
				meta: meta,
			}, nil
		},
		"db": func() (cli.Command, error) {
			return &dbCommand{
				meta: meta,
			}, nil
		},
		"db delete": func() (cli.Command, error) {
			return &dbDeleteCommand{
				meta: meta,
			}, nil
		},
		"db validate": func() (cli.Command, error) {
			return &dbValidateCommand{
				meta: meta,
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		meta.ui.Error(err.Error())
	}
	os.Exit(exitStatus)

}
