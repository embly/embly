package nixbuild

import (
	nixbuildpb "embly/pkg/nixbuild/pb"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type BuildServiceServer struct {
}

func (service *BuildServiceServer) Build(server nixbuildpb.BuildService_BuildServer) error {
	return nil
}

func (b *Builder) startServer() (err error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9276))
	if err != nil {
		return
	}
	b.server = grpc.NewServer()
	nixbuildpb.RegisterBuildServiceServer(b.server, &BuildServiceServer{})
	return b.server.Serve(lis)
}
