package vinyl

import (
	fmt "fmt"
	"reflect"
	"testing"

	"github.com/embly/vinyl/vinyl-go/transport"
	proto "github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func TestAllMethod(t *testing.T) {
	// db := DB{}

	qs := []transport.Query{}
	if err := fillInterfaceWithType(&qs); err != nil {
		t.Error(err)
	}
	fmt.Println(qs)
	// if err := db.All(&qs); err != nil {
	// 	t.Error(err)
	// }
}

func fillInterfaceWithType(msgs interface{}) (err error) {
	v := reflect.ValueOf(msgs)
	if v.Kind() != reflect.Ptr {
		return errors.Errorf("must be passed a pointer to a slice %v", v.Type())
	}
	v = v.Elem()

	if v.Kind() != reflect.Slice {
		return errors.Errorf("can't fill non-slice value")
	}

	v.Set(reflect.MakeSlice(v.Type(), 3, 3))
	for i := 0; i < 3; i++ {
		proto.Unmarshal(
			[]byte{10, 8, 119, 104, 97, 116, 101, 118, 101, 114, 18, 11, 109, 97, 120, 64, 109, 97, 120, 46, 99, 111, 109},
			v.Index(i).Addr().Interface().(proto.Message),
		)
	}
	return nil
}

func marshalAndReturn() (out []transport.Query) {
	out = make([]transport.Query, 3)
	for i := 0; i < 3; i++ {
		proto.Unmarshal(
			[]byte{10, 8, 119, 104, 97, 116, 101, 118, 101, 114, 18, 11, 109, 97, 120, 64, 109, 97, 120, 46, 99, 111, 109},
			&out[i],
		)
	}
	return
}
func BenchmarkReflect(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		qs := []transport.Query{}
		if err := fillInterfaceWithType(&qs); err != nil {
			b.Error(err)
		}
	}
}
func BenchmarkTyped(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		marshalAndReturn()
	}
}
