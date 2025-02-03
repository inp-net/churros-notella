ARG TAG=dev

FROM registry.inpt.fr/inp-net/images/go-just:1.23.5-1.39.0 AS builder

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
RUN just build notella ${TAG}

# Stage 2: Create a lightweight image with just the binary
FROM alpine:3.21 AS runner

# Set the working directory in the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/notella /app/notella

# Command to run the binary
CMD ["/app/notella"]
 
