package dock

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"embly/pkg/jsonmessage"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/segmentio/textio"
)

// NewClient creates and returns a new dock client, erroring if it can't connect to docker
func NewClient() (client *Client, err error) {
	client = &Client{ctx: context.Background()}
	if client.client, err = dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithVersion("1.38")); err != nil {
		err = errors.WithStack(err)
	}
	return
}

// Client is the dock client, a docker client specific to our use cases
type Client struct {
	client *client.Client
	ctx    context.Context
}

// DownloadImageIfStaleOrUnavailable takes an image name, displays download logs to stdout.
// Checking for stale is "TODO"
func (c *Client) DownloadImageIfStaleOrUnavailable(image string) (err error) {
	exists, err := c.ImageExists(image)
	if err != nil {
		return
	}
	// TODO: check if it's the latest, and maybe ignore the error if there's no
	// internet connectivity and we have a local stale image
	if !exists {
		fmt.Printf("image '%s' not found locally, downloading\n", image)
		return c.PullImage(image)
	}
	return
}

// ImageExists checks if an image exists locally
func (c *Client) ImageExists(image string) (exists bool, err error) {
	if _, _, err = c.client.ImageInspectWithRaw(c.ctx, image); dockerclient.IsErrNotFound(err) {
		return false, nil
	} else if err != nil {
		err = errors.WithStack(err)
		return
	}
	exists = true
	return
}

// VolumeExists checks if a volume exists
func (c *Client) VolumeExists(name string) (exists bool, err error) {
	if _, err = c.client.VolumeInspect(c.ctx, name); dockerclient.IsErrNotFound(err) {
		return false, nil
	} else if err != nil {
		return
	}
	exists = true
	return
}

// CreateVolume create a local volume with no tags and the default settings
func (c *Client) CreateVolume(name string) (err error) {
	_, err = c.client.VolumeCreate(c.ctx, volume.VolumeCreateBody{
		Driver:     "local",
		DriverOpts: map[string]string{},
		Labels:     map[string]string{},
		Name:       name,
	})
	return
}

// ImageCreated returns the time the image was created on the host machine
func (c *Client) ImageCreated(image string) (created time.Time, err error) {
	inspect, _, err := c.client.ImageInspectWithRaw(c.ctx, image)
	if err != nil {
		return
	}
	created, err = time.Parse(time.RFC3339, inspect.Created)
	return
}

// PullImage pulls an image by name, pretty-prints the download status to stdout
func (c *Client) PullImage(image string) (err error) {
	readCloser, err := c.client.ImagePull(c.ctx, image, types.ImagePullOptions{})
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if err = jsonmessage.DisplayJSONMessagesStream(readCloser, os.Stdout, os.Stdout.Fd(), true, nil); err != nil {
		err = errors.WithStack(err)
		return
	}
	readCloser.Close()
	return
}

// NewContainer created a new instance of a Container struct
func (c *Client) NewContainer(name, image string) *Container {
	return &Container{
		Name:   name,
		Image:  image,
		client: c,
		Binds:  make(map[string]string),
		Ports:  make(map[string]string),
	}
}

// Container is a docker container
type Container struct {
	Name       string
	WorkingDir string
	Binds      map[string]string
	Ports      map[string]string
	Cmd        []string
	Image      string
	client     *Client
	ExecPrefix string
}

// Create creates the docker container, make sure you have added all necessary configuration
// values before running
func (c *Container) Create() (err error) {
	var binds []string
	for k, v := range c.Binds {
		binds = append(binds, fmt.Sprintf("%s:%s", k, v))
	}

	portMap := make(nat.PortMap)
	portSet := make(nat.PortSet)
	for k, v := range c.Ports {
		// TODO: support docker compose functionality here
		portMap[nat.Port(k)] = []nat.PortBinding{{
			HostIP:   "127.0.0.1",
			HostPort: v,
		}}
		portSet[nat.Port(k)] = struct{}{}
	}

	if _, err = c.client.client.ContainerCreate(c.client.ctx, &container.Config{
		Image:        c.Image,
		Cmd:          strslice.StrSlice(c.Cmd),
		ExposedPorts: portSet,
		WorkingDir:   c.WorkingDir,
	}, &container.HostConfig{
		Binds:        binds,
		PortBindings: portMap,
	},
		&network.NetworkingConfig{}, c.Name); err != nil {
		err = errors.WithStack(err)
	}
	return
}

// Stop the docker container
func (c *Container) Stop() (err error) {
	timeout := time.Second
	return c.client.client.ContainerStop(c.client.ctx, c.Name, &timeout)
}

// Start the docker container
func (c *Container) Start() (err error) {
	err = c.client.client.ContainerStart(c.client.ctx, c.Name, types.ContainerStartOptions{})
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}

// Remove the docker container
func (c *Container) Remove() (err error) {
	return c.client.client.ContainerRemove(c.client.ctx, c.Name, types.ContainerRemoveOptions{})
}

// Exec a command inside the docker container
func (c *Container) Exec(cmd string) (err error) {
	cli := c.client.client
	ctx := c.client.ctx

	var execID types.IDResponse
	if execID, err = cli.ContainerExecCreate(ctx, c.Name, types.ExecConfig{
		Cmd:          []string{"bash", "-c", cmd},
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
		textio.NewPrefixWriter(os.Stdout, c.ExecPrefix),
		textio.NewPrefixWriter(os.Stderr, c.ExecPrefix),
		hr.Reader)
	hr.Close()
	if err != nil {
		return
	}

	var execInspect types.ContainerExecInspect
	if execInspect, err = cli.ContainerExecInspect(ctx, execID.ID); err != nil {
		return
	}
	if execInspect.ExitCode != 0 {
		err = errors.Errorf("Got non zero exit code for exec: %d", execInspect.ExitCode)
	}
	return

}

// Copy files or directories from place to place, attempts to mirror functionality of docker cp
// src is CONTAINER:SRC_PATH or SRC_PATH
// dst is CONTAINER:DEST_PATH or DEST_PATH
func (c *Client) Copy(src, dest string) (err error) {

	// TODO: fully implement following
	// https://docs.docker.com/engine/reference/commandline/cp/

	srcParts := strings.Split(src, ":")
	var toCopy io.ReadCloser
	var isDir bool = true
	if len(srcParts) > 1 {
		if toCopy, _, err = c.client.CopyFromContainer(c.ctx, srcParts[0], srcParts[1]); err != nil {
			return
		}
	} else {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)

		src, err = filepath.Abs(src)
		if err != nil {
			return err
		}

		var srcInfo os.FileInfo
		srcInfo, err = os.Stat(src)
		if err != nil {
			return err
		}
		copyLocationBase := filepath.Base(src)
		isDir = srcInfo.IsDir()

		if err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			name, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			if !isDir {
				name = srcInfo.Name()
			}

			if info.IsDir() && !isDir {
				return nil
			}
			if isDir {
				name = filepath.Join(copyLocationBase, name)
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = name

			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return
		}
		if err = tw.Close(); err != nil {
			return
		}
		toCopy = ioutil.NopCloser(&buf)
	}

	destParts := strings.Split(dest, ":")
	if len(destParts) > 1 {
		destLoc := destParts[1]
		if !isDir {
			destLoc = filepath.Dir(destLoc)
		}
		if err = c.client.CopyToContainer(
			c.ctx, destParts[0], destLoc,
			toCopy, types.CopyToContainerOptions{
				AllowOverwriteDirWithFile: true,
			}); err != nil {
			return
		}
	} else {
		tr := tar.NewReader(toCopy)
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break // End of archive
			}
			if err != nil {
				return err
			}
			target := filepath.Join(dest, header.Name)
			if header.Typeflag == tar.TypeDir {
				if _, err := os.Stat(target); err == nil {
					continue
				}
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			} else if header.Typeflag == tar.TypeReg {
				f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return errors.WithStack(err)
				}
				if _, err := io.Copy(f, tr); err != nil {
					return err
				}
				f.Close()
			}
		}
	}
	toCopy.Close()
	return
}

// CopyFile copes a file from within the docker container to the host os. Currently does no validation on
// the src and dst values, using with directories will error or be weird
func (c *Container) CopyFile(src, dest string) (err error) {

	cli := c.client.client
	ctx := c.client.ctx
	var tarWasmOut io.ReadCloser
	if tarWasmOut, _, err = cli.CopyFromContainer(ctx, c.Name, src); err != nil {
		return
	}
	tr := tar.NewReader(tarWasmOut)
	wasmFile, err := os.Create(dest)
	if err != nil {
		return
	}
	_, err = tr.Next()
	if err != nil {
		return err
	}
	f, err := os.Create(wasmFile.Name())
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, tr); err != nil {
		return err
	}

	_, err = tr.Next()
	if err != io.EOF {
		return errors.New("tried to copy one file but multiple files were found")
	}
	err = nil

	return
}
