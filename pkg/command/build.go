package command

import (
	"embly/pkg/build"
	"strings"
)

func runBuild(m *meta) (co build.CompileOutput, er int) {
	ef, err := build.FindAndValidateEmblyFile()
	if err != nil {
		m.ui.Error(err.Error())
		er = 1
		return
	}

	co, err = build.CompileFunctions(ef, m.ui)
	if err != nil {
		m.ui.Error(err.Error())
		er = 1
		return
	}
	return
}

type buildCommand struct {
	*meta
}

func (f *buildCommand) Help() string {
	return strings.TrimSpace(`
Usage: embly build (<function-name>)...

    Build a local embly project. 
	`)
}

func (f *buildCommand) Run(args []string) int {
	_, er := runBuild(f.meta)
	return er
}

func (f *buildCommand) Synopsis() string {
	return "Build an embly project"
}
