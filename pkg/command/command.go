package command

import (
	"fmt"
	"os"
	"strings"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
	flag "github.com/spf13/pflag"
)

type errRunResultHelp struct {
}

func (e *errRunResultHelp) Error() string {
	return "run result help error"
}

type command interface {
	flags() *flag.FlagSet
	help() string
	run(args []string) (err error)
	synopsis() string
}

type wrapper struct {
	command command
}

func (w *wrapper) Help() string {
	f := w.command.flags()
	var options string
	if f != nil {
		options = fmt.Sprintf("\n\nOptions:\n%s", f.FlagUsages())
	}
	return strings.TrimSpace(
		w.command.help() + options,
	)
}
func (w *wrapper) Run(args []string) int {
	f := w.command.flags()
	if f != nil {
		_ = f.Parse(args)
		args = f.Args()
	}
	if err := w.command.run(args); err != nil {
		if _, ok := err.(*errRunResultHelp); ok {
			return cli.RunResultHelp
		}
		UI.Error(err.Error())
		fmt.Printf("%+v", err)
		return 1
	}
	return 0
}

func (w *wrapper) Synopsis() string {
	return w.command.synopsis()
}

// UI is the command UI
var UI cli.Ui

func createCLI() *cli.CLI {
	c := cli.NewCLI("embly", "0.0.1")
	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"dev":         wrap(&devCommand{}),
		"run":         wrap(&runCommand{}),
		"bundle":      wrap(&bundleCommand{}),
		"build":       wrap(&buildCommand{}),
		"db":          factory(&dbCommand{}),
		"db delete":   wrapSimple(dbDeleteCommand),
		"db validate": wrapSimple(dbValidateCommand),
	}
	return c
}

func init() {
	UI = &cli.ColoredUi{
		Ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
		InfoColor: cli.UiColor{
			Bold: true,
		},
	}
}

// Run the embly command line
func Run() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "embly",
		Output: &cli.UiWriter{Ui: UI},
	})
	_ = logger
	c := createCLI()
	exitStatus, err := c.Run()
	if err != nil {
		UI.Error(err.Error())
	}
	os.Exit(exitStatus)

}

func factory(c cli.Command) cli.CommandFactory {
	return func() (cli.Command, error) {
		return c, nil
	}
}
func wrap(c command) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &wrapper{command: c}, nil
	}
}
func wrapSimple(s simple) cli.CommandFactory {
	return wrap(&s)
}
