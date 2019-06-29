package rustcompile

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/gin-gonic/gin"
	rc "embly/api/pkg/rustcompile/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Start ...
func Start(port int) string {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpc.WithInsecure()
	grpcServer := grpc.NewServer()
	rc.RegisterRustCompileServer(grpcServer, &server{})

	go grpcServer.Serve(lis)

	return lis.Addr().String()
}

type server struct{}

func (s *server) StartBuild(code *rc.Code, stream rc.RustCompile_StartBuildServer) error {
	result := &rc.Result{Log: "Hello"}
	if err := stream.Send(result); err != nil {
		return err
	}
	return errors.New("hmmm")
}

func indexHandler(ctx context.Context, db *sql.DB, c *gin.Context) error {
	c.JSON(200, gin.H{"msg": "Hello"})
	return nil
}
