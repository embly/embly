package descriptor

import (
	"bytes"
	"compress/gzip"
	fmt "fmt"
	"io/ioutil"

	proto "github.com/golang/protobuf/proto"
	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/pkg/errors"
)

func ptrString(val string) *string {
	return &val
}
func ptrInt32(val int32) *int32 {
	return &val
}

// AddRecordTypeUnion takes a compressed protobuf descriptor and adds the RecordTypeUnion
func AddRecordTypeUnion(descriptorBytes []byte, records []string) (out []byte, err error) {
	zr, err := gzip.NewReader(bytes.NewBuffer(descriptorBytes))
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(zr)
	if err != nil {
		return
	}
	zr.Close()
	fdp := pb.FileDescriptorProto{}
	// TODO add validation. It seems like we can be passed avalid descriptor with no records.
	// at the very least confirm that there are records for each record name passed
	if err = proto.Unmarshal(b, &fdp); err != nil {
		err = errors.Wrap(err, "error unmarshaling the descriptor, is it the right format?")
		return
	}
	descriptor := pb.DescriptorProto{
		Name: ptrString("RecordTypeUnion"),
	}
	var packageName string
	if fdp.Package != nil {
		packageName = "." + *fdp.Package
	}
	for i, record := range records {
		field := pb.FieldDescriptorProto{
			Name:     ptrString("_" + record),
			Number:   ptrInt32(int32(i + 1)),
			Label:    pb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			Type:     pb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
			TypeName: ptrString(fmt.Sprintf("%s.%s", packageName, record)),
			JsonName: &record,
		}
		descriptor.Field = append(descriptor.Field, &field)
	}

	// TODO: check if it is already part of the proto definition?
	// ...might actually be fine if it's duplicated
	fdp.MessageType = append(fdp.MessageType, &descriptor)

	return proto.Marshal(&fdp)
}
