FROM alpine:latest
RUN apk --no-cache add ca-certificates


ENTRYPOINT ["/gordon-proxy"]