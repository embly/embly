package nixbuild

import "google.golang.org/grpc"

import nixbuildpb "embly/pkg/nixbuild/pb"

import "context"

func (b *Builder) connectToBuildServer() (err error) {
	conn, err := grpc.Dial("localhost:9276")
	if err != nil {
		return
	}
	b.client = nixbuildpb.NewBuildServiceClient(conn)
	// health endpoint check?

	return nil
}

func (b *Builder) startRemoteBuild(name string) (err error) {
	buildClient, err := b.client.Build(context.Background())
	if err != nil {
		return
	}

	buildClient.Send(&nixbuildpb.ClientPayload{
		Build: &nixbuildpb.Build{
			Name:          name,
			BuildLocation: loc,
		},
	})
}
