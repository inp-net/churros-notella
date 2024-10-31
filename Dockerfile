# Stage 1: Build the Go binary using Just
FROM golang:1.20 AS builder

# Install Just in the builder stage
RUN apt-get update && apt-get install -y curl && \
    curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to /usr/local/bin/just

# Set the working directory in the container
WORKDIR /app

# Copy the go.mod and go.sum files first to cache the dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the Justfile and source code
COPY Justfile .
COPY . .

# Build the Go binary using Just
RUN just build

# Stage 2: Create a lightweight image with just the binary
FROM alpine:latest

# Set the working directory in the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/myapp .

# Expose a port (optional)
EXPOSE 8080

# Command to run the binary
CMD ["./myapp"]
