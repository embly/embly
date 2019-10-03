package dock

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
)

// https://github.com/FoundationDB/fdb-record-layer/blob/master/docs/SchemaEvolution.md#validating-changes

// RunVinylPrefix the prefix added to the vinyl container name
var RunVinylPrefix = "embly-vinyl-"

// VinylImage is the docker image used for vinyl
var VinylImage = "embly/vinyl:latest"

// Vinyl is an instance of a vinyl database
type Vinyl struct {
	Cont *Container
	Port int
	Name string
}

func getNewRandomPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port, nil
}

func (v *Vinyl) Wait() (err error) {
	reader, err := v.Cont.client.client.ContainerLogs(
		v.Cont.client.ctx, RunVinylPrefix+v.Name, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
	if err != nil {
		return err
	}
	defer reader.Close()

	r, w := io.Pipe()
	go stdcopy.StdCopy(w, w, reader)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "Server started") {
			return nil
		}
	}
	return scanner.Err()
}

// StartVinyl starts a database instance of vinyl
func StartVinyl(name string) (vinyl *Vinyl, err error) {
	// find a free port
	port, err := getNewRandomPort()
	if err != nil {
		return
	}
	vinyl = &Vinyl{
		Name: name,
		Port: port,
	}
	c, err := NewClient()
	if err != nil {
		return
	}

	if err = c.DownloadImageIfStaleOrUnavailable(VinylImage); err != nil {
		return
	}

	volumeName := RunVinylPrefix + name + "-volume"

	if err = c.CreateVolume(volumeName); err != nil {
		return
	}

	vinyl.Cont = c.NewContainer(RunVinylPrefix+name, VinylImage)
	vinyl.Cont.Binds[volumeName] = "/var/fdb/data"
	vinyl.Cont.Ports["8090"] = strconv.Itoa(port)

	// remove any existing containers and ignore the error
	// TODO: bad?
	vinyl.Cont.Stop()
	vinyl.Cont.Remove()

	if err = vinyl.Cont.Create(); err != nil {
		return
	}

	// TODO: wait until "Server started" has been seen in the logs? how?

	err = vinyl.Cont.Start()
	return
}
