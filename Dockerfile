ARG TAG=dev

# Stage 1: Build the Go binary using Just
FROM golang:1.23.4-alpine3.20 AS builder

# Install Just in the builder stage
RUN apk add --no-cache curl bash git
RUN curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to /app

ENV PATH="/app:${PATH}"

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
RUN /app/just build notella ${TAG}

# Stage 2: Create a lightweight image with just the binary
FROM alpine:3.20 AS runner

# Set the working directory in the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/notella /app/notella

# Command to run the binary
CMD ["/app/notella"]
 
