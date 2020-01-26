package nixbuild

//go:generate protoc -I . ./nixbuild.proto --go_out=pb=grpc:pb
//go:generate statik -include=*.nix -src=../../compilers/nix/
