package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

type dbCommand struct {
	*meta
}

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
	return "Run a local embly project"
}

type dbDeleteCommand struct {
	*meta
}

func (f *dbDeleteCommand) Help() string {
	return strings.TrimSpace(`
Usage: embly db delete <name>

    Delete all records in an embly database
	`)
}
func (f *dbDeleteCommand) Run(args []string) int {
	return cli.RunResultHelp
}
func (f *dbDeleteCommand) Synopsis() string {
	return "Delete all records in an embly database"
}

type dbValidateCommand struct {
	*meta
}

func (f *dbValidateCommand) Help() string {
	return strings.TrimSpace(`
Usage: embly db validate <name>

    Validate a new database schema against an existing db 
	`)
}
func (f *dbValidateCommand) Run(args []string) int {
	return cli.RunResultHelp
}
func (f *dbValidateCommand) Synopsis() string {
	return "Validate a new db schema"
}
