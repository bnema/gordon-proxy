.PHONY: build-arm64 podman-build

build-arm64:
	@env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -o bin/gordon-proxy-arm64-bin
	@file bin/gordon-proxy-arm64-bin

podman-build: build-arm64
	@podman build -t gordon-proxy:latest .

podman-export: podman-build
	@podman save -o bin/gordon-proxy-latest.tar gordon-proxy:latest

all: podman-build podman-export