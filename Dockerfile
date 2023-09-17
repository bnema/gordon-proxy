FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY ./bin/gordon-proxy-arm64-bin /gordon-proxy-arm64-bin
CMD ["/gordon-proxy-arm64-bin"]
