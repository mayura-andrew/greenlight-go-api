# Dockerfile
FROM ubuntu:20.04

# Install build essentials and the required GLIBC version
RUN apt-get update && apt-get install -y build-essential

# Install a specific GLIBC version
# Replace 'GLIBC_VERSION' with the version from your server
RUN apt-get install -y libc6=2.31

# Copy your source code
COPY . /app

# Set working directory
WORKDIR /app

# Build the binary
RUN make build/api
