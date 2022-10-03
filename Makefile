IMG ?= controller:latest

.PHONY: build
build:
	go build

.PHONY: docker-build
docker-build:
	docker build -t ${IMG} .
