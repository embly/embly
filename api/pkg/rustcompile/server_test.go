package rustcompile

// import (
// 	"context"
// 	"fmt"
// 	"testing"

// 	rc "embly/api/pkg/rustcompile/proto"

// 	"google.golang.org/grpc"
// )

// func TestStart(t *testing.T) {
// 	addr := Start(0)
// 	conn, err := grpc.Dial(addr, grpc.WithInsecure())
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	defer conn.Close()

// 	client := rc.NewRustCompileClient(conn)
// 	buildClient, err := client.StartBuild(context.Background(), &rc.Code{})
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	r, err := buildClient.Recv()
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	fmt.Println(r)

// 	r, err = buildClient.Recv()
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	fmt.Println(r)

// }
