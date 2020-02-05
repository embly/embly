SHELL = /usr/bin/env bash

BUILDDIR = build
OBJDIR = ./build


frontend/node_modules: frontend/package.json
	cd frontend && yarn

build_blog: $(OBJDIR)/build_blog
$(OBJDIR)/build_blog: blog/content/* blog/templates/* blog/sass/* blog/static/*
	cd blog && ./build.sh
	touch build/build_blog

build_frontend: $(OBJDIR)/build_frontend
$(OBJDIR)/build_frontend: frontend/src/* frontend/package.json frontend/public/* frontend/node_modules
	cd frontend && yarn build
	touch build/build_frontend

embly_bundle: build_frontend build_blog
	embly bundle

build_docker_image: embly_bundle
	docker-compose build app

push_docker_image: build_docker_image
	docker --config ~/.docker-embly push embly/app:latest

run:
	cd .. && make build_embly
	embly dev
