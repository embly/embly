package core

//go:generate protoc -I ./proto ./proto/comms.proto --go_out=plugins=grpc:proto
//go:generate protoc -I ./httpproto ./httpproto/http.proto --go_out=plugins=grpc:httpproto
