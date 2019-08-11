package localbuild

import (
	"archive/tar"
	"context"
	"embly/pkg/randy"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"
	"github.com/segmentio/textio"
	"github.com/sirupsen/logrus"
)

var imageName = "maxmcd/embly-compile-rust-wasm"

// Create ...
func Create(fName, buildLocation, buildContext, destination string) (err error) {
	var cli *client.Client
	if cli, err = client.NewClientWithOpts(client.FromEnv); err != nil {
		return
	}
	fmt.Println("building with: ", buildLocation, buildContext, destination)

	ctx := context.Background()

	tmpDir := fmt.Sprintf("/tmp/%s", randy.String())
	containerName := "embly-rust-build" + fName

	if _, err = cli.ContainerCreate(ctx, &container.Config{
		Image: "maxmcd/embly-compile-rust-wasm:clean",
		Cmd:   strslice.StrSlice([]string{"sleep", "1000"}),
	}, &container.HostConfig{
		Binds: []string{buildContext + ":/opt/context"},
	},
		&network.NetworkingConfig{}, containerName); err != nil {
		if !strings.Contains(err.Error(), "is already in use by container") {
			return
		}
	}

	defer func() {
		timeout := time.Millisecond * 1
		if err := cli.ContainerStop(ctx, containerName, &timeout); err != nil {
			return
		}
		// if err = cli.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{}); err != nil {
		// 	return
		// }
	}()

	if err = cli.ContainerStart(ctx, containerName, types.ContainerStartOptions{}); err != nil {
		return
	}

	if err = execInContainerAndWait(ctx, cli, containerName, []string{"mkdir", "-p", tmpDir}); err != nil {
		return
	}

	if err = cli.CopyToContainer(ctx, containerName, tmpDir, nil, types.CopyToContainerOptions{}); err != nil {
		return
	}
	var relLocation string
	if relLocation, err = filepath.Rel(buildContext, buildLocation); err != nil {
		return
	}

	outName := filepath.Base(destination)
	if err = execInContainerAndWait(ctx, cli, containerName, []string{"bash", "-c", fmt.Sprintf(`
set -x
cd %s \
&& ls -lah \
&& mkdir -p /opt/out \
&& rm /opt/out/*.wasm || true \
&& cargo +nightly build --target wasm32-wasi --release -Z unstable-options --out-dir /opt/out \
&& wasm-strip /opt/out/*.wasm \
&& ls -lah /opt/out/*.wasm \
&& mv /opt/out/*.wasm /opt/out/%s || true
	`, filepath.Join("/opt/context", relLocation), outName)}); err != nil {
		err = errors.WithStack(err)
		return
	}

	var tarWasmOut io.ReadCloser
	if tarWasmOut, _, err = cli.CopyFromContainer(ctx, containerName, "/opt/out/"+outName); err != nil {
		return
	}

	tr := tar.NewReader(tarWasmOut)

	tmpWasmFile, err := ioutil.TempFile("", "embly-wasm-out")
	if err != nil {
		return
	}
	defer os.Remove(tmpWasmFile.Name())
	// todo: here should only be one file here, ensure that is the case
	for {
		_, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		f, err := os.Create(tmpWasmFile.Name())
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			return err
		}
	}

	bindingsLocation, err := writeBindingsFile()
	defer os.Remove(bindingsLocation)
	if err != nil {
		fmt.Println("asdfasdfasasdfd")
		return err
	}

	err = runLucetc(bindingsLocation, tmpWasmFile.Name(), destination)
	if err != nil {
		fmt.Println("asdfasdfasd")
		return err
	}

	return
}

func execInContainerAndWait(ctx context.Context, cli *client.Client, containerName string, cmd []string) (err error) {
	logrus.Debug("Running command : ", cmd, " in container ", containerName)
	var execID types.IDResponse
	if execID, err = cli.ContainerExecCreate(ctx, containerName, types.ExecConfig{
		Cmd:          cmd,
		Tty:          true,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
	}); err != nil {
		return
	}

	var hr types.HijackedResponse
	if hr, err = cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{Detach: false, Tty: false}); err != nil {
		return
	}
	_, err = stdcopy.StdCopy(
		textio.NewPrefixWriter(os.Stdout, "stdout: "),
		textio.NewPrefixWriter(os.Stderr, "stderr: "),
		hr.Reader)
	return
}
