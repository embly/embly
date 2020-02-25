module embly

go 1.12

replace github.com/docker/docker v0.0.0-20170601211448-f5ec1e2936dc => github.com/docker/engine v0.0.0-20190822180741-9552f2b2fdde

require (
	github.com/bytecodealliance/lucet v0.0.0-20200110144859-4b5916182acc // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.0.0-20170601211448-f5ec1e2936dc
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0
	github.com/embly/vinyl v0.0.0-20191002235719-de628fc27e36
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/hashicorp/go-hclog v0.10.1
	github.com/hashicorp/hcl2 v0.0.0-20191002203319-fb75b3253c80
	github.com/mitchellh/cli v1.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/morikuni/aec v1.0.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/radovskyb/watcher v1.0.7
	github.com/rakyll/statik v0.1.7-0.20200207095328-0c7347ad9ff9
	github.com/segmentio/textio v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/zclconf/go-cty v1.2.0
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553 // indirect
	golang.org/x/sys v0.0.0-20191224085550-c709ea063b76 // indirect
	golang.org/x/tools v0.0.0-20191227053925-7b8e75db28f4
	google.golang.org/grpc v1.26.0
)
