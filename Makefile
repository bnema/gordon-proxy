.PHONY: build-arm64

build-arm64:
	@env GOOS=linux GOARCH=arm64 go build -o gordon-proxy-arm64
	@file gordon-proxy-arm64

