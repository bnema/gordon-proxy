.PHONY: build-arm64 podman-build

build-arm64:
	@env GOOS=linux GOARCH=arm64 go build -o bin/gordon-proxy-arm64-bin
	@file bin/gordon-proxy-arm64-bin

podman-build: build-arm64
	@podman build -t gordon-proxy:latest .

podman-export: podman-build
	@podman save -o bin/gordon-proxy-latest.tar gordon-proxy:latest

all: build-arm64 podman-build podman-export