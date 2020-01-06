package core

//go:generate protoc -I ./proto ./proto/comms.proto --go_out=plugins=grpc:proto
//go:generate protoc -I ./http_proto ./http_proto/http.proto --go_out=plugins=grpc:http_proto
