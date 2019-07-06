package rustcompile

import (
	rc "embly/api/pkg/rustcompile/proto"

	"google.golang.org/grpc"
)

func NewRustCompileClient(target string) (rcc rc.RustCompileClient, err error) {
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return
	}
	rcc = rc.NewRustCompileClient(conn)
	return
}
