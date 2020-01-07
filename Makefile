SHELL = /usr/bin/env bash

BUILDDIR = build

OBJDIR = ./build

install_embly_wrapper: $(OBJDIR)/install_embly_wrapper
$(OBJDIR)/install_embly_wrapper: embly-wrapper-rs/* embly-wrapper-rs/src/*
	cd embly-wrapper-rs && cargo build --release
	touch build/install_embly_wrapper

install_embly: $(OBJDIR)/install_embly
$(OBJDIR)/install_embly: cmd/embly/* pkg/**/* go.mod go.sum
	cd cmd/embly && go install
	touch build/install_embly

generate_http_proto: $(OBJDIR)/generate_http_proto
$(OBJDIR)/generate_http_proto: pkg/core/httpproto/http.proto
	cd pkg/core && go generate
	cd embly-rs && ./gen_proto.sh
	touch build/generate_http_proto

generate_comms_proto: $(OBJDIR)/generate_comms_proto
$(OBJDIR)/generate_comms_proto: pkg/core/proto/comms.proto
	cd pkg/core && go generate
	cd embly-wrapper-rs && cargo build
	touch build/generate_comms_proto

all: 
	make -j install_embly install_embly_wrapper generate_http_proto generate_comms_proto

test:
	make -j wrapper_test cargo_test go_test

go_test:
	go test ./... -cover

wrapper_test:
	cd embly-wrapper-rs && cargo test

cargo_test: 
	cargo test

doc_test:
	cargo test --doc

install_rust_toolchain:	
	rustup toolchain add nightly-2019-11-24
	rustup target add wasm32-wasi --toolchain nightly-2019-11-24 

run_mjpeg_example: build
	cd examples/mjpeg && embly dev

run_kv_example: build
	cd examples/kv && embly dev

bundle_project_example: build
	cd examples/project && embly bundle

run_project_example: build
	cd examples/project && embly dev

clean:
	rm build/*
