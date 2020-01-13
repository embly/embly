package dock

import (
	"embly/pkg/filesystem"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
)

// CompileRustSettings settings for rust compilation
type CompileRustSettings struct {
	FunctionName   string
	BuildLocation  string
	ProjectRoot    string
	DestinationDir string
	Sources        []string
}

// CompileRustPrefix is the prefix added to the container name
var CompileRustPrefix = "embly-rust-build-"

// CompileRustImage is the docker image used to compile
var CompileRustImage = "embly/compile-rust-wasm:slim"

func (settings *CompileRustSettings) containerName() string {
	// return CompileRustPrefix + settings.FunctionName
	return CompileRustPrefix + "-all"
}

// CompileRust starts a docker container, bind mounts a volume with the appropriate source
// files, compiles them into wasm with cargo and copies the resulting wasm file onto
// the host machine
func CompileRust(settings CompileRustSettings) (err error) {
	c, err := NewClient()
	if err != nil {
		return
	}

	if err = c.DownloadImageIfStaleOrUnavailable(CompileRustImage); err != nil {
		return
	}

	cont := c.NewContainer(settings.containerName(), CompileRustImage)

	cont.Cmd = []string{"sleep", "100000"}
	cont.ExecPrefix = fmt.Sprintf("[%s]:", settings.FunctionName)
	if err = cont.Create(); err != nil {
		_ = err
		// TODO: assert on "already exists" error and ignore
		// return
	}
	if err = cont.Start(); err != nil {
		return
	}

	newBuildLocation, archive, err := filesystem.ZipSources(settings.ProjectRoot, settings.BuildLocation, settings.Sources)
	if err != nil {
		return
	}

	_ = cont.Exec("mkdir -p /opt/context") // ignore error
	_ = cont.Exec("rm -rf /opt/context/*") // ignore error

	if err = cont.client.client.CopyToContainer(
		cont.client.ctx, cont.Name, "/opt/context", archive,
		types.CopyToContainerOptions{
			AllowOverwriteDirWithFile: true,
		}); err != nil {
		err = errors.WithStack(err)
		return
	}

	var sb strings.Builder
	sb.WriteString("cd ")
	sb.WriteString(filepath.Join("/opt/context/", newBuildLocation))
	// -Zno-index-update
	sb.WriteString(" && cargo +nightly build --target wasm32-wasi --release -Z unstable-options --out-dir /opt/out")
	buildCommand := sb.String()

	defer cont.Stop()
	_ = cont.Exec("mkdir -p /opt/out")               // ignore error
	_ = cont.Exec("rm /opt/out/*.wasm 2> /dev/null") // ignore error
	err = cont.Exec(buildCommand)
	if err != nil {
		return
	}
	err = cont.Exec("wasm-strip /opt/out/*.wasm")
	if err != nil {
		err = errors.Wrap(err, "error running wasm-strip")
		return
	}

	_ = cont.Exec(fmt.Sprintf(
		"mv /opt/out/*.wasm /opt/out/%s 2> /dev/null",
		settings.FunctionName+".wasm")) // ignore error

	wasmFile := filepath.Join(settings.DestinationDir, settings.FunctionName+".wasm")
	return cont.CopyFile("/opt/out/"+settings.FunctionName+".wasm", wasmFile)
}
