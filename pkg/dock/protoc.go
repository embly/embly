package dock

import (
	"embly/pkg/randy"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	proto "github.com/golang/protobuf/proto"
	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// ProtocImage is the protoc docker image
const ProtocImage = "embly/protoc"

// DescriptorForFile takes a local filepath, downloads the protoc docker container
// if it's not already downloaded, copies the single file into the protoc container
// generates a descriptor and returns the raw bytes of the descriptor
func DescriptorForFile(file string) (descriptor []byte, err error) {
	c, err := NewClient()
	if err != nil {
		return
	}

	if err = c.DownloadImageIfStaleOrUnavailable(ProtocImage); err != nil {
		return
	}
	name := randy.String()
	cont := c.NewContainer(name, ProtocImage)
	cont.Cmd = []string{"sleep", "10000"}
	if err = cont.Create(); err != nil {
		return
	}

	if err = cont.Start(); err != nil {
		return
	}
	defer func() {
		cont.Stop()
		cont.Remove()
	}()
	if err = c.Copy(file, name+":/opt/"); err != nil {
		return
	}

	if err = cont.Exec(fmt.Sprintf(
		"protoc -o /tmp/foo -I /opt /opt/%s",
		filepath.Base(file))); err != nil {
		return
	}

	d, err := ioutil.TempDir("", "")
	if err != nil {
		return
	}
	defer os.RemoveAll(d)

	if err = c.Copy(name+":/tmp/foo", d); err != nil {
		return
	}
	descriptorSetBytes, err := ioutil.ReadFile(filepath.Join(d, "foo"))
	if err != nil {
		return
	}
	descriptorSet := pb.FileDescriptorSet{}
	if err = proto.Unmarshal(descriptorSetBytes, &descriptorSet); err != nil {
		return
	}
	return proto.Marshal(descriptorSet.File[0])
}
