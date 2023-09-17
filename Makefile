.PHONY: build-arm64

build-arm64:
	@env GOOS=linux GOARCH=arm64 go build -o bin/gordon-proxy-arm64-bin
	@file bin/gordon-proxy-arm64-bin

