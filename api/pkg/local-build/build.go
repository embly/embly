package localbuild

import (
	"archive/tar"
	"bytes"
	"context"
	"embly/api/pkg/randy"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
)

func create() (err error) {
	var cli *client.Client
	if cli, err = client.NewClientWithOpts(client.FromEnv); err != nil {
		return
	}

	tmpDir := fmt.Sprintf("/tmp/%s", randy.String())
	ctx := context.Background()
	cfg := container.Config{
		Image: "maxmcd/embly-compile-rust-wasm",
		Cmd:   strslice.StrSlice([]string{"sleep", "10"}),
	}
	hostConfig := container.HostConfig{}
	networkingConfig := network.NetworkingConfig{}
	containerName := "embly-rust-build" + randy.String()
	var ccBody container.ContainerCreateCreatedBody

	if ccBody, err = cli.ContainerCreate(ctx, &cfg, &hostConfig, &networkingConfig, containerName); err != nil {
		return
	}

	defer func() {
		timeout := time.Millisecond * 1
		if err = cli.ContainerStop(ctx, ccBody.ID, &timeout); err != nil {
			return
		}
		if err = cli.ContainerRemove(ctx, ccBody.ID, types.ContainerRemoveOptions{}); err != nil {
			return
		}
	}()

	if err = cli.ContainerStart(ctx, ccBody.ID, types.ContainerStartOptions{}); err != nil {
		return
	}

	var execID types.IDResponse
	if execID, err = cli.ContainerExecCreate(ctx, ccBody.ID, types.ExecConfig{
		Cmd: []string{"mkdir", "-p", tmpDir},
	}); err != nil {
		return
	}
	fmt.Println(execID)
	if err = cli.ContainerExecStart(ctx, execID.ID, types.ExecStartCheck{}); err != nil {
		return
	}

	for {
		var info types.ContainerExecInspect
		if info, err = cli.ContainerExecInspect(ctx, execID.ID); err != nil {
			return
		}
		if info.Running != true {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}

	toCopy := filesTar()
	if err = cli.CopyToContainer(ctx, ccBody.ID, tmpDir,
		&toCopy, types.CopyToContainerOptions{}); err != nil {
		return
	}
	fmt.Println(ccBody)

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
