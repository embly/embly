package nixbuild

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"embly/pkg/filesystem"
	nixbuildpb "embly/pkg/nixbuild/pb"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type BuildServiceServer struct {
	builder *Builder
}

func (service *BuildServiceServer) Build(server nixbuildpb.BuildService_BuildServer) (err error) {
	defer func() {
		fmt.Printf("Server returned with error: %+v\n", err)
		err = errors.WithStack(err)
	}()
	payload, err := server.Recv()
	if payload.Build == nil {
		return errors.New("first payload must include build details")
	}
	respPay := &nixbuildpb.ServerPayload{
		HashNeeded: []*nixbuildpb.HashRequest{},
	}
	loc := payload.Build.BuildLocation
	name := payload.Build.Name
	fileMap := map[string]*nixbuildpb.File{}
	neededFiles := map[string]struct{}{}
	for _, file := range payload.Build.Files {
		fileMap[file.Path] = file
		if file.IsDir {
			continue
		}
		if file.Size == 0 {
			continue
		}
		if !service.builder.hashedFileExists(file.Hash) {
			respPay.HashNeeded = append(respPay.HashNeeded, &nixbuildpb.HashRequest{
				Hash: file.Hash,
				Path: file.Path,
			})
			neededFiles[file.Path] = struct{}{}
		}
	}
	if err = server.Send(respPay); err != nil {
		return
	}
	// sent off our file requests
	for {
		if len(neededFiles) == 0 {
			// we have all the files we need
			break
		}
		payload, err = server.Recv()
		fmt.Println("SERVER: ", payload)
		if err != nil {
			return
		}
		var thisHash []byte
		// the hash may have changed under our feet if edits happened during an upload
		// so recompute the hash and update it if we need to
		thisHash, err = service.builder.writeFileToBlobCache(payload.File)
		fmt.Println(err)
		if err != nil {
			return
		}
		fileReference := fileMap[payload.File.Path]
		if bytes.Equal(fileReference.Hash, payload.File.RequestedHash) {
			fileReference.Hash = thisHash
		} else {
			return errors.Errorf(
				"file with path %s has hash %x but client sent hash %x, failing build",
				payload.File.Path, fileReference.Hash, payload.File.RequestedHash)
		}
		delete(neededFiles, payload.File.Path)
	}

	dir, err := service.builder.constructNetworkedBuild(loc, fileMap)
	if err != nil {
		return
	}
	fmt.Println("DIRECTORY", dir)
	result, err := service.builder.BuildDirectory(dir, name)
	if err != nil {
		return
	}
	fmt.Println(result)
	return
}

func (b *Builder) constructNetworkedBuild(loc string, fileMap map[string]*nixbuildpb.File) (buildDir string, err error) {
	// using symlinks seems fun and lightweight, maybe it's a bad idea
	defer func() {
		err = errors.WithStack(err)
	}()

	fileList := []string{}
	for name := range fileMap {
		fileList = append(fileList, name)
	}
	prefix := filesystem.CommonPrefix(fileList)
	sort.Strings(fileList)
	buildDir, err = ioutil.TempDir(b.emblyLoc("./build"), "embly-build")
	if err != nil {
		return
	}
	for _, file := range fileList {
		fi := fileMap[file]
		newLoc := filepath.Join(buildDir, strings.TrimPrefix(file, prefix))
		if fi.IsDir {
			if err = os.MkdirAll(newLoc, os.ModeDir|os.ModePerm); err != nil {
				return
			}
		} else {
			if err = os.Symlink(
				b.emblyLoc("./blob_cache/"+fmt.Sprintf("%x", fi.Hash)),
				newLoc); err != nil {
				return
			}
		}
	}
	buildDir = filepath.Join(buildDir, strings.TrimPrefix(loc, prefix))
	return
}

func (b *Builder) writeFileToBlobCache(file *nixbuildpb.CompressedFile) (hash []byte, err error) {
	defer func() {
		err = errors.WithStack(err)
	}()
	tmpFile, err := ioutil.TempFile(b.emblyLoc("./blob_cache/"), "prehash")
	if err != nil {
		return
	}
	compressedReader, err := zlib.NewReader(bytes.NewBuffer(file.Body))
	if err != nil {
		return
	}
	hashing := sha256.New()
	// use a teeReader to ioCopy to the tmpFile and the sha256
	// compute the hash while writing to the file
	tee := io.TeeReader(compressedReader, hashing)

	// EOF is ok, just use the hash of the empty file
	// TODO: optimize this away later?
	if _, err = io.Copy(tmpFile, tee); err != nil {
		err = errors.WithStack(err)
		return
	}
	if err = errors.WithStack(compressedReader.Close()); err != nil {
		return
	}
	if err = tmpFile.Close(); err != nil {
		return
	}
	hash = hashing.Sum(nil)
	err = errors.WithStack(os.Rename(
		tmpFile.Name(),
		b.emblyLoc("./blob_cache/"+fmt.Sprintf("%x", hash))))
	return
}

func (b *Builder) hashedFileExists(hash []byte) (exists bool) {
	_, err := os.Stat(b.emblyLoc("./blob_cache/" + fmt.Sprintf("%x", hash)))
	return err == nil
}

func (b *Builder) startServer() (err error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9276))
	if err != nil {
		return
	}
	b.server = grpc.NewServer()
	nixbuildpb.RegisterBuildServiceServer(b.server, &BuildServiceServer{
		builder: b,
	})
	return b.server.Serve(lis)
}
