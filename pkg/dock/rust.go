package dock

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// CompileRustSettings settings for rust compilation
type CompileRustSettings struct {
	FunctionName   string
	BuildLocation  string
	BuildContext   string
	DestinationDir string
}

// CompileRustPrefix is the prefix added to the container name
var CompileRustPrefix = "embly-rust-build-"

// CompileRustImage is the docker image used to compile
var CompileRustImage = "embly/compile-rust-wasm:slim"

func (settings *CompileRustSettings) containerName() string {
	return CompileRustPrefix + settings.FunctionName
}

func (settings *CompileRustSettings) buildCommand() (out string, err error) {
	var sb strings.Builder
	sb.WriteString("cd ")
	relativeLocation, err := filepath.Rel(settings.BuildContext, settings.BuildLocation)
	if err != nil {
		errors.Wrap(err, "Error finding the relative project path from the build context")
		return
	}
	sb.WriteString(filepath.Join("/opt/context", relativeLocation))
	sb.WriteString(" && cargo +nightly build --target wasm32-wasi --release -Z unstable-options --out-dir /opt/out")
	out = sb.String()
	return
}

// CompileRust starts a docker container, bind mounts a volume with the appropriate source
// files, compiles them into wasm with cargo and copies the resulting wasm file onto
// the host machine
func CompileRust(settings CompileRustSettings) (err error) {
	buildCommand, err := settings.buildCommand()
	if err != nil {
		return
	}

	c, err := NewClient()
	if err != nil {
		return
	}

	if err = c.DownloadImageIfStaleOrUnavailable(CompileRustImage); err != nil {
		return
	}

	cont := c.NewContainer(settings.containerName(), CompileRustImage)

	// TODO, this will never be updated if the values change, just errors on create
	// and then uses the stale container
	cont.Binds = map[string]string{
		settings.BuildContext: "/opt/context",
	}
	cont.Cmd = []string{"sleep", "100000"}
	cont.ExecPrefix = fmt.Sprintf("[%s]:", settings.FunctionName)
	if err = cont.Create(); err != nil {

		// return
	}
	if err = cont.Start(); err != nil {
		return
	}
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
