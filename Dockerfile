# Use scratch for the minimal base image
FROM scratch

# Copy the pre-built binary from your host system
COPY ./bin/gordon-proxy-arm64-bin /gordon-proxy-arm64-bin

# Command to run the application
CMD ["/gordon-proxy-arm64-bin"]