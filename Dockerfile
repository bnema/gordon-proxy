FROM alpine:latest
RUN apk --no-cache add ca-certificates bash


COPY gordon-proxy /

ENTRYPOINT ["/gordon-proxy"]