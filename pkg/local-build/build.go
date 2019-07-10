package localbuild

import (
	"archive/tar"
	"bytes"
	"context"
	"embly/api/pkg/randy"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/segmentio/textio"
	"github.com/sirupsen/logrus"
)

func Create() (err error) {
	var cli *client.Client
	if cli, err = client.NewClientWithOpts(client.FromEnv); err != nil {
		return
	}

	tmpDir := fmt.Sprintf("/tmp/%s", randy.String())
	ctx := context.Background()
	cfg := container.Config{
		Image: "maxmcd/embly-compile-rust-wasm",
		Cmd:   strslice.StrSlice([]string{"sleep", "100"}),
	}
	hostConfig := container.HostConfig{}
	networkingConfig := network.NetworkingConfig{}
	containerName := "embly-rust-build"

	if _, err = cli.ContainerCreate(ctx, &cfg, &hostConfig, &networkingConfig, containerName); err != nil {
		if !strings.Contains(err.Error(), "\"/embly-rust-build\" is already in use by container") {
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

	toCopy := filesTar()
	if err = cli.CopyToContainer(ctx, containerName, tmpDir,
		&toCopy, types.CopyToContainerOptions{}); err != nil {
		return
	}
	fmt.Println("hi")
	if err = execInContainerAndWait(ctx, cli, containerName, []string{"bash", "-c", fmt.Sprintf(`
cd %s \
&& cargo +nightly build --target wasm32-wasi --release -Z unstable-options --out-dir ./out \
&& wasm-strip ./out/*.wasm \
&& ls -lah ./out/*.wasm
	`, tmpDir)}); err != nil {
		return
	}
	var tarWasmOut io.ReadCloser
	if tarWasmOut, _, err = cli.CopyFromContainer(ctx, containerName, tmpDir+"/out/foo.wasm"); err != nil {
		return
	}

	tr := tar.NewReader(tarWasmOut)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		f, err := os.Create(hdr.Name)
		if err != nil {
			return err
		}
		fmt.Printf("Contents of %s:\n", hdr.Name)
		if _, err := io.Copy(f, tr); err != nil {
			return err
		}
		fmt.Println()
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

func filesTar() (buf bytes.Buffer) {
	tw := tar.NewWriter(&buf)
	var files = []struct {
		Name, Body string
	}{
		{"src/main.rs", `fn main(){ println!("hi") }`},
		{"Cargo.toml", "[package]\nname = \"foo\"\nversion = \"0.0.1\"\n[dependencies]\nembly=\"0.0.2\""},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}
	return
}
