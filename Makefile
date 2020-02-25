SHELL = /usr/bin/env bash

BUILDDIR = build

OBJDIR = ./build

install_embly_wrapper: $(OBJDIR)/install_embly_wrapper
$(OBJDIR)/install_embly_wrapper: embly-wrapper-rs/* embly-wrapper-rs/src/*
	cd embly-wrapper-rs && cargo install --force --path .
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

build_build_server: $(OBJDIR)/build_build_server
$(OBJDIR)/build_build_server: pkg/nixbuild/* pkg/nixbuild/**/* cmd/build-server/*
	cd cmd/build-server \
	&& docker -l=info build -f ./Dockerfile -t embly/build-server ../.. \
	&& docker kill embly-build || true \
	&& docker rm embly-build || true
	touch build/build_build_server

all:
	make -j install_embly install_embly_wrapper generate_http_proto generate_comms_proto

generate: generate_comms_proto generate_http_proto
	cd pkg/nixbuild && go generate

build_embly:
	make -j install_embly install_embly_wrapper

build_embly_image:
	cd nix && ./build-embly-embly.sh

push_embly_image:
	docker --config ~/.docker-embly push embly/embly

test:
	make -j wrapper_test cargo_test go_test

build_server_shell:
	docker run -it embly/build-server sh

push_build_server: build_build_server
	docker --config=~/.docker-embly  push embly/build-server

build_examples: build_embly
	cd examples/mjpeg && embly build
	cd examples/kv && embly build
	cd examples/project && embly build

build_hello_world: install_embly build_build_server
	cd examples/hello-world && embly build

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

run_mjpeg_example: build_embly
	cd examples/mjpeg && embly dev

run_kv_example: build_embly
	cd examples/kv && embly dev

bundle_kv_example: build_embly
	cd examples/kv && embly bundle

bundle_project_example: build_embly
	cd examples/project && embly bundle

new_build: install_embly
	cd examples/hello-world && embly build

run_project_example: build_embly
	cd examples/project && embly dev

clean:
	rm build/*

deploy_embly_run_no_embly_image:
	cd app && make push_docker_image
	./tools/deploy_embly_run.sh

deploy_embly_run: build_embly_image
	cd app && make push_docker_image
	./tools/deploy_embly_run.sh

embly_run_logs:
	./tools/embly_run_logs.sh

build_blog_examples: build_hello_world
	cd examples/hello-world && ./inject_example.py
	cd app && make build_blog
	cd examples/hello-world && ./copy_example_html.py
