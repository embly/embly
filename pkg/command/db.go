package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

type dbCommand struct{}

func (f *dbCommand) Help() string {
	return strings.TrimSpace(`
Usage: embly db <command> <name>

    Run various database maintenace tasks. 
	`)
}
func (f *dbCommand) Run(args []string) int {
	return cli.RunResultHelp
}
func (f *dbCommand) Synopsis() string {
	return "Run various database maintenace tasks. "
}

var dbDeleteCommand = simple{
	synopsisVal: "Delete all records in an embly database",
	helpVal: `Usage: embly db delete <name>

	Delete all records in an embly database`,
	runFunc: func(args []string) error {
		return &errRunResultHelp{}
	},
}

var dbValidateCommand = simple{
	synopsisVal: "Validate a new db schema",
	helpVal: `
	Usage: embly db validate <name>
	
		Validate a new database schema against an existing db 
		`,
	runFunc: func(args []string) error {
		return &errRunResultHelp{}
	},
}
