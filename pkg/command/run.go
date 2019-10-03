package command

import (
	"embly/pkg/core"
	"fmt"
	"strings"
)

type runCommand struct {
	*meta
}

func (f *runCommand) Help() string {
	return strings.TrimSpace(`
Usage: embly run

    Run a local embly project. 
	`)
}

func (f *runCommand) Run(args []string) int {
	co, er := runBuild(f.meta)
	if er != 0 {
		return er
	}
	_ = co

	if len(co.Config.Gateways) == 0 {
		f.ui.Info("No gateways, nothing to run")
		return 0
	}
	err := core.Start(co, f.ui)
	if err != nil {
		fmt.Printf("%+v\n", err)
		f.ui.Error(err.Error())
		return 1
	}
	return 0
}
func (f *runCommand) Synopsis() string {
	return "Run a local embly project"
}
