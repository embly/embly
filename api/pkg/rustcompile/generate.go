package rustcompile

//go:generate protoc -I ./proto ./proto/rustcompile.proto --go_out=plugins=grpc:proto
