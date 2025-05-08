# Use the official Go image as a builder
FROM public.ecr.aws/docker/library/golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first, then download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go binary
RUN go build -o server ./cmd/server

# Use a minimal image for the final stage
FROM alpine:latest

# Install CA certificates (needed for HTTPS)
RUN apk --no-cache add ca-certificates

# Set working directory in final image
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Command to run the binary
CMD ["./server"]
