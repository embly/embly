package build

import (
	"archive/tar"
	"compress/gzip"
	"embly/pkg/config"
	"embly/pkg/dock"
	"embly/pkg/filesystem"
	"embly/pkg/lucet"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"
	"golang.org/x/tools/godoc/vfs"

	"github.com/mitchellh/cli"
)

// Builder build projects
type Builder struct {
	Config    *config.Config
	ui        cli.Ui
	Functions map[string]Files
}

func (builder *Builder) emblyBuildDir() string {
	return filepath.Join(builder.Config.ProjectRoot, "embly_build")
}

func NewBuilderFromArchive(path string, ui cli.Ui) (builder *Builder, err error) {
	builder = &Builder{}
	f, err := os.Open(path)
	if err != nil {
		err = errors.New("archive location doesn't exist")
		return
	}

	projectRoot, err := ioutil.TempDir("", "")
	if err != nil {
		return
	}
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			return builder, err
		}

		// the target location where the dir/file should be created
		target := filepath.Join(projectRoot, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return builder, err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return builder, err
			}
			if _, err := io.Copy(f, tr); err != nil {
				return builder, err
			}
			f.Close()
		}
	}

	return NewBuilder(projectRoot, ui)
}

// NewBuilder returns any errors reading and validating the configuration
func NewBuilder(path string, ui cli.Ui) (builder *Builder, err error) {
	builder = &Builder{ui: ui, Functions: make(map[string]Files)}
	builder.Config, err = config.New(path)
	return
}

// Files tracks the location of specific output files
type Files struct {
	Wasm string
	Obj  string
}

func (builder *Builder) WatchForChangesAndRebuild() (err error) {
	w := watcher.New()

	watching := make([]string, len(builder.Config.Functions))
	for i, fn := range builder.Config.Functions {
		location := filepath.Join(builder.Config.ProjectRoot, fn.Path)
		watching[i] = location
		if err := w.AddRecursive(location); err != nil {
			return err
		}
		for _, s := range fn.Sources {
			if err := w.AddRecursive(s); err != nil {
				return err
			}
		}
	}

	isBuilding := make([]bool, len(watching))
	shouldBuild := make(chan int, 100)
	go func() {
		for i := range shouldBuild {
			if !isBuilding[i] {
				isBuilding[i] = true
				builder.ui.Info("rebuilding function")
				go func() {
					if err := builder.build(builder.Config.Functions[i]); err != nil {
						fmt.Printf("%+v", err)
						builder.ui.Error(err.Error())
						return
					}
					isBuilding[i] = false
					builder.ui.Info("rebuilding complete")
				}()
				// build
			}
		}
	}()
	go func() {
		for {
			select {
			case event := <-w.Event:
				for i := 0; i < len(watching); i++ {
					if strings.HasPrefix(event.Path, watching[i]) {
						shouldBuild <- i
					}
				}
			case err := <-w.Error:
				fmt.Println("error watching files", err)
			case <-w.Closed:
				return
			}
		}
	}()
	go w.Start(time.Millisecond * 300)
	return nil
}

func (builder *Builder) initBuildDirectory() (err error) {
	emblyBuildDir := builder.emblyBuildDir()
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
	return nil
}

// CompileFunctions compiles completely and return any related errors
func (builder *Builder) CompileFunctions() (err error) {
	if err := builder.initBuildDirectory(); err != nil {
		return err
	}
	for _, fn := range builder.Config.Functions {
		if err := builder.build(fn); err != nil {
			return err
		}
	}
	return
}

// CompileWasmToObject just takes wasm files and turns them into local object files
func (builder *Builder) CompileWasmToObject(isTar bool) (err error) {
	// TODO: support no building with a tar?
	if isTar {

		for _, fn := range builder.Config.Functions {
			objLocation := filepath.Join(builder.emblyBuildDir(), fn.Name+"."+builder.objectExtension())
			_, err := os.Stat(objLocation)
			if err == nil {
				builder.addObjFile(fn.Name, objLocation)
			}
		}
		return nil
	}

	if err := builder.initBuildDirectory(); err != nil {
		return err
	}
	for _, fn := range builder.Config.Functions {
		if err := builder.compileWasm(fn); err != nil {
			return err
		}
	}
	return
}

// CompileFunctionsToWasm just outputs project wasm files for all functions
func (builder *Builder) CompileFunctionsToWasm() (err error) {
	if err := builder.initBuildDirectory(); err != nil {
		return err
	}
	for _, fn := range builder.Config.Functions {
		if err := builder.compileFunction(fn); err != nil {
			return err
		}
	}
	return
}

func (builder *Builder) Bundle(location string, includeObjectFiles bool) (err error) {
	fs := filesystem.FileSystem{FileSystem: vfs.OS(builder.Config.ProjectRoot)}
	archive, _, err := fs.Bundle(includeObjectFiles)
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(archive)
	if err != nil {
		return
	}
	return ioutil.WriteFile(location, b, 0644)
}

func (builder *Builder) addWasmFile(name, loc string) {
	files := builder.Functions["function."+name]
	files.Wasm = loc
	builder.Functions["function."+name] = files

}

func (builder *Builder) addObjFile(name, loc string) {
	files := builder.Functions["function."+name]
	files.Obj = loc
	builder.Functions["function."+name] = files

}

func (builder *Builder) objectExtension() string {
	return runtime.GOOS
}

func (builder *Builder) compileWasm(fn config.Function) (err error) {
	builder.ui.Output(fmt.Sprintf("Compiling %s.wasm to a local object file", fn.Name))
	start := time.Now()
	wasmFile := filepath.Join(builder.emblyBuildDir(), fn.Name+".wasm")
	objLocation := filepath.Join(builder.emblyBuildDir(), fn.Name+"."+builder.objectExtension())
	if err = lucet.CompileWasmToObject(wasmFile, objLocation); err != nil {
		return
	}

	builder.ui.Info(fmt.Sprintf("Compilation of function '%s' complete (%s)", fn.Name, time.Now().Sub(start)))
	builder.addObjFile(fn.Name, objLocation)
	return
}

func (builder *Builder) compileFunction(fn config.Function) (err error) {
	builder.ui.Info(fmt.Sprintf("Building function '%s'", fn.Name))
	builder.ui.Output(fmt.Sprintf(`Compiling "%s"`, fn.Name))
	if err = dock.CompileRust(dock.CompileRustSettings{
		FunctionName:   fn.Name,
		Sources:        fn.Sources,
		BuildLocation:  fn.Path,
		ProjectRoot:    builder.Config.ProjectRoot,
		DestinationDir: builder.emblyBuildDir(),
	}); err != nil {
		return
	}
	wasmFile := filepath.Join(builder.emblyBuildDir(), fn.Name+".wasm")
	builder.addWasmFile(fn.Name, wasmFile)
	return
}

func (builder *Builder) build(fn config.Function) (err error) {
	if err := builder.compileFunction(fn); err != nil {
		return err
	}
	return builder.compileWasm(fn)
}
