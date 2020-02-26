package nixbuild

import (
	"embly/pkg/config"
	"embly/pkg/embly"
	"embly/pkg/filesystem"
	"embly/pkg/lucet"
	nixbuildpb "embly/pkg/nixbuild/pb"
	_ "embly/pkg/nixbuild/statik"
	"fmt"
	"io/ioutil"
	"strings"

	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
)

type Builder struct {
	ui cli.Ui

	emblyDir string
	project  *filesystem.Project

	server *grpc.Server
	client nixbuildpb.BuildServiceClient

	useLocalNix bool
}

// NixCommandExists tries to run the nix command to see if it exists in the system
func NixCommandExists() bool {
	cmd := exec.Command("nix")
	_, err := cmd.CombinedOutput()
	return err == nil
}

// NewClientBuilder creates a new builder, but prompts the user if it's unclear
// if a build server should be used or local nix
func NewClientBuilder(ui cli.Ui) (b *Builder, err error) {
	b, err = NewBuilder()
	b.ui = ui

	if b.checkForBuildServer() {
		return
	}

	nixExists := NixCommandExists()
	// check for linux...
	// maybe with the nix command?
	if nixExists {
		var resp string
		resp, err = b.ui.Ask("Would you like to use nix to build?")
		if err != nil {
			return
		}
		if strings.Contains(resp, "y") {
			b.useLocalNix = true
			return
		}
	}

	var resp string
	resp, err = b.ui.Ask("Would you like to start a build server?")
	if err != nil {
		return
	}
	if strings.Contains(resp, "y") {
		if err = b.startDockerServer(); err != nil {
			return
		}
		return
	}
	err = errors.New("unable to start")
	return
}

// NewBuildServer creates a new server that's intended to be run as a server
func NewBuildServer() (b *Builder, err error) {
	b, err = NewBuilder()
	if err != nil {
		return
	}
	if err = b.writeNixFiles(); err != nil {
		return
	}
	return
}

func NewBuilder() (b *Builder, err error) {
	b = &Builder{}
	if err = embly.CreateHomeDir(); err != nil {
		return
	}
	b.emblyDir, err = embly.EmblyDir()
	return
}

func (b *Builder) SetProject(cfg *config.Config) {
	b.project = filesystem.NewProject(cfg)
}

func (b *Builder) emblyLoc(path string) string {
	return filepath.Join(b.emblyDir, path)
}

func (b *Builder) CleanAllDependencies() (err error) {
	dirs, err := filepath.Glob(b.emblyLoc("./nix/") + "*.*")
	if err != nil {
		return
	}
	for _, file := range dirs {
		if err = os.RemoveAll(file); err != nil {
			return
		}
	}
	return
}

func (b *Builder) Build(name string) (result string, err error) {
	if b.useLocalNix {
		return b.BuildFunction(name)
	} else {
		return b.startRemoteBuild(name)
	}
}

func (b *Builder) BuildDirectory(dir, functionName string, logWriter io.Writer) (result string, err error) {
	result, err = ioutil.TempDir(b.emblyLoc("./result/"), "")
	if err != nil {
		return
	}
	{
		cmd := exec.Command(
			"nix", "run", "-i",
			"-f", b.emblyLoc("./nix/rust-run.nix"),
			"--keep", "CARGO_TARGET_DIR",
			"--keep", "CARGO_HOME",
			"-c",
			"cargo", "build", "--release", "--target=wasm32-wasi",
			"-Z", "unstable-options", "--out-dir", result,
		)
		cmd.Dir = dir
		cmd.Env = []string{
			// this is needed for `nix run` it doesn't get passed into the command context
			fmt.Sprintf("NIX_PATH=%s", os.Getenv("NIX_PATH")),

			fmt.Sprintf("CARGO_TARGET_DIR=%s", b.emblyLoc("./build_context/rust_target")),
			fmt.Sprintf("CARGO_HOME=%s", b.emblyLoc("./build_context/cargo_home")),
		}
		cmd.Stderr = logWriter
		cmd.Stdout = logWriter

		if err = cmd.Run(); err != nil {
			return
		}
	}
	{
		var loc string
		loc, err = lucet.WriteBindingsFile()
		if err != nil {
			return
		}
		cmd := exec.Command(
			"nix", "run", "-i",
			"-f", b.emblyLoc("./nix/wasm-run.nix"),
			"--keep", "LDFLAGS",
			"--keep", "LD",
			"-c",
			"sh", "-c",
			`set -x && wasm-strip *.wasm \
			&& lucetc --bindings `+loc+`\
			--emit so \
			--opt-level 0 *.wasm \
			-o `+functionName+`.out \
			&& strip `+functionName+`.out`,
		)
		cmd.Dir = result
		cmd.Env = []string{
			// this is needed for `nix run` it doesn't get passed into the command context
			fmt.Sprintf("NIX_PATH=%s", os.Getenv("NIX_PATH")),
			// 	"LDFLAGS=-dylib -dead_strip -export_dynamic -undefined dynamic_lookup",
			// 	"LD=/nix/store/qlqy7maq14k4wxlc16yjikf12x5b0dln-cctools/bin/ld",
		}
		cmd.Stderr = logWriter
		cmd.Stdout = logWriter

		if err = cmd.Run(); err != nil {
			return
		}
	}
	return
}

func (b *Builder) BuildFunction(name string) (result string, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	if b.project == nil {
		err = errors.New("no project initialized, no function to run")
		return
	}

	dir, err := b.project.CopyFunctionSourcesToBuild(b.emblyLoc("./build"), name)
	if err != nil {
		return
	}
	defer os.RemoveAll(dir)
	return b.BuildDirectory(dir, name, os.Stdout)
}

func (b *Builder) writeNixFiles() (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()

	staticFiles, err := fs.New()
	if err != nil {
		return
	}
	if err = fs.Walk(
		staticFiles, "/",
		func(path string, fi os.FileInfo, e error) (err error) {
			defer func() {
				err = errors.WithStack(err)
			}()

			if e != nil || fi.IsDir() {
				return e
			}
			file, err := staticFiles.Open(path)
			if err != nil {
				return
			}
			to, err := os.OpenFile(b.emblyLoc("./nix/"+fi.Name()),
				os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				return
			}
			defer to.Close()
			if _, err = io.Copy(to, file); err != nil {
				return
			}
			return
		},
	); err != nil {
		return
	}
	return
}

func (b *Builder) DownloadDependencies(items []string) (err error) {
	for _, item := range items {
		if _, err = os.Stat(b.emblyLoc("./nix/" + item)); err == nil {
			// directory exists, continue to other dependencies
			continue
		}
		if err = b.BuildFile(item); err != nil {
			return
		}
	}
	return
}

func (b *Builder) BuildFile(name string) (err error) {
	defer func() {
		err = errors.WithStack(err)
	}()

	staticFiles, err := fs.New()
	if err != nil {
		return
	}
	f, err := staticFiles.Open(fmt.Sprintf("/%s.nix", name))
	if err != nil {
		return
	}

	cmd := exec.Command("nix-build", "-", "-o", name)
	emblyDir, err := embly.EmblyDir()
	if err != nil {
		return
	}
	cmd.Dir = filepath.Join(emblyDir, "./nix")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	if _, err = io.Copy(stdin, f); err != nil {
		return
	}
	stdin.Close()

	if err = cmd.Start(); err != nil {
		return
	}
	if err = cmd.Wait(); err != nil {
		return
	}

	return nil

}
