# Start from scratch for the final image
FROM scratch


COPY gordon-proxy /

ENTRYPOINT ["/gordon-proxy"]
