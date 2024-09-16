# Stage 1: Build the Go application
FROM golang:1.17-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o watermark main.go

# Stage 2: Create a minimal image with the built binary
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/watermark .

# Expose the port that the app listens on
EXPOSE 8080

# Command to run when starting the container
ENTRYPOINT ["./watermark"]
