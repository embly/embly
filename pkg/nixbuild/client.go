package nixbuild

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"context"
	nixbuildpb "embly/pkg/nixbuild/pb"
)

func (b *Builder) checkForBuildServer() bool {
	conn, err := grpc.Dial(
		"localhost:9276",
		grpc.WithInsecure(),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: time.Millisecond * 300,
		}))
	if err != nil {
		fmt.Println(err)
		return false
	}
	b.client = nixbuildpb.NewBuildServiceClient(conn)
	fmt.Println(b.serverHealthy())
	return b.serverHealthy() == nil
}

func (b *Builder) serverHealthy() (err error) {
	if b.client == nil {
		return errors.New("client doesn't exist")
	}
	resp, err := b.client.Health(context.Background(), &nixbuildpb.HealthPayload{})
	if err != nil {
		return err
	}
	if resp.Code < 300 {
		return nil
	}
	return errors.Errorf(
		`got unhealthy build server response code %d with message "%s"`,
		resp.Code, resp.Msg)
}
func (b *Builder) connectToBuildServer() (err error) {
	if err = b.serverHealthy(); err != nil {
		return
	}
	conn, err := grpc.Dial("localhost:9276", grpc.WithInsecure())
	if err != nil {
		return
	}
	b.client = nixbuildpb.NewBuildServiceClient(conn)
	return b.serverHealthy()
}

func WriteCompressedFile(file *nixbuildpb.CompressedFile, dir string) (err error) {
	// Consider streaming in chunks?
	r, err := zlib.NewReader(bytes.NewBuffer(file.Body))
	if err != nil {
		return
	}
	to, err := os.OpenFile(filepath.Join(dir, file.Name), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if _, err = io.Copy(to, r); err != nil {
		err = errors.WithStack(err)
		return
	}
	return errors.WithStack(r.Close())
}

func ReadCompressedFile(loc string) (compFile *nixbuildpb.CompressedFile, err error) {
	var b bytes.Buffer
	writer := zlib.NewWriter(&b)
	var f *os.File
	f, err = os.Open(loc)
	if err != nil {
		return
	}
	if _, err = io.Copy(writer, f); err != nil {
		return
	}
	if err = writer.Close(); err != nil {
		return
	}
	compFile = &nixbuildpb.CompressedFile{
		Body: b.Bytes(),
		Path: loc,
		Name: filepath.Base(loc),
	}
	return
}

func (b *Builder) startRemoteBuild(name string) (result string, err error) {
	defer func() {
		fmt.Println("Client returned with error: ", err)
		err = errors.WithStack(err)
	}()
	buildClient, err := b.client.Build(context.Background())
	if err != nil {
		return
	}

	path, files, err := b.project.FunctionWatchedFiles(name)
	if err != nil {
		return
	}
	var protoFiles []*nixbuildpb.File
	for loc, file := range files {
		if err = file.PopulateHash(loc); err != nil {
			return
		}
		protoFiles = append(protoFiles, &nixbuildpb.File{
			Path:  loc,
			Name:  file.Name(),
			IsDir: file.IsDir(),
			Size:  file.Size(),
			Hash:  file.Hash,
		})
	}

	if err = buildClient.Send(&nixbuildpb.ClientPayload{
		Build: &nixbuildpb.Build{
			Name:          name,
			BuildLocation: path,
			Files:         protoFiles,
		},
	}); err != nil {
		return
	}
	var pay *nixbuildpb.ServerPayload
	for {
		pay, err = buildClient.Recv()
		if err != nil {
			return
		}
		for _, hashRequest := range pay.HashNeeded {
			if _, ok := files[hashRequest.Path]; !ok {
				panic(fmt.Sprint(hashRequest.Path, "is unknown to this client, panicking!"))
			}
			var compFile *nixbuildpb.CompressedFile
			compFile, err = ReadCompressedFile(hashRequest.Path)
			if err != nil {
				return
			}
			compFile.RequestedHash = hashRequest.Hash
			if err = buildClient.Send(&nixbuildpb.ClientPayload{
				File: compFile,
			}); err != nil {
				return
			}
		}
		for _, log := range pay.Log {
			// prefixwriter?
			os.Stdout.Write(log)
		}

		if len(pay.Files) > 0 {
			result, err = ioutil.TempDir(b.emblyLoc("./result/"), "")
			if err != nil {
				return
			}
			for _, file := range pay.Files {
				fmt.Printf("Writing output file %s to %s\n", file.Name, result)
				if err = WriteCompressedFile(file, result); err != nil {
					return
				}
			}
			break // we got our files, build is complete
		}
	}
	return
}
