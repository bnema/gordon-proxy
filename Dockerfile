# Use a base Alpine Linux image for ARM64
FROM arm64v8/alpine:latest

# Install any necessary dependencies if required
RUN apk update && apk add --no-cache ...

# Copy the ARM64 binary into the container
COPY gordon-proxy /usr/local/bin/

# Set the entry point to run your binary
ENTRYPOINT ["/usr/local/bin/gordon-proxy"]

