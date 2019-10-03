package build

import (
	"embly/pkg/config"
	"embly/pkg/dock"
	"embly/pkg/lucet"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/mitchellh/cli"
)

// EmblyFile holds a configuration and the location of the project
type EmblyFile struct {
	Config      config.Config
	ProjectRoot string
}

func (ef *EmblyFile) emblyBuildDir() string {
	return filepath.Join(ef.ProjectRoot, "embly_build")
}

// FindAndValidateEmblyFile returns any errors reading and validating the configuration
func FindAndValidateEmblyFile() (ef EmblyFile, err error) {
	f, l, err := config.FindConfigFile()
	if err != nil {
		return
	}
	ef.Config, err = config.ParseConfig(f)
	ef.ProjectRoot = l
	return
}

// CompileOutput holds resulting build metadata
type CompileOutput struct {
	EmblyFile
	Functions map[string]Files
}

// Files tracks the location of specific output files
type Files struct {
	Wasm string
	Obj  string
}

// CompileFunctions and return any related errors
func CompileFunctions(ef EmblyFile, ui cli.Ui) (co CompileOutput, err error) {
	co.Functions = make(map[string]Files)
	co.EmblyFile = ef

	emblyBuildDir := ef.emblyBuildDir()
	ebFileInfo, _ := os.Stat(emblyBuildDir)
	if ebFileInfo == nil {
		if err = os.Mkdir(emblyBuildDir, os.ModePerm); err != nil {
			err = errors.WithStack(err)
			return
		}
	} else {
		if !ebFileInfo.IsDir() {
			err = errors.New("embly_build exists but it is not a directory")
			return
		}
	}

	for _, fn := range ef.Config.Functions {
		ui.Info(fmt.Sprintf("Building function '%s'", fn.Name))
		buildContext := filepath.Join(ef.ProjectRoot, fn.Context)
		ui.Output(fmt.Sprintf(`Compiling "%s" with context "%s" in location "%s"`, fn.Name, buildContext, filepath.Join(buildContext, fn.Path)))
		if err = dock.CompileRust(dock.CompileRustSettings{
			FunctionName:   fn.Name,
			BuildContext:   buildContext,
			BuildLocation:  filepath.Join(buildContext, fn.Path),
			DestinationDir: emblyBuildDir,
		}); err != nil {
			return
		}

		ui.Output(fmt.Sprintf("Compiling %s.wasm to a local object file", fn.Name))
		wasmFile := filepath.Join(emblyBuildDir, fn.Name+".wasm")
		objLocation := filepath.Join(emblyBuildDir, fn.Name+".out")
		if err = lucet.CompileWasmToObject(wasmFile, objLocation); err != nil {
			return
		}
		co.Functions["function."+fn.Name] = Files{
			Wasm: wasmFile,
			Obj:  objLocation,
		}

		ui.Info(fmt.Sprintf("Compilation of function '%s' complete", fn.Name))
	}

	return
}
