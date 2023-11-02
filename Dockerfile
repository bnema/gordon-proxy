FROM alpine

COPY gordon-proxy /

ENTRYPOINT ["/gordon-proxy"]
