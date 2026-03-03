# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (no longer need C compiler for PostgreSQL)
RUN apk add --no-cache git

# Enable automatic toolchain download for newer Go versions
ENV GOTOOLCHAIN=auto

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (no CGO needed for PostgreSQL driver)
RUN GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Set environment variable for JWT secret (can be overridden)
ENV JWT_SECRET="development-secret-change-in-production"

# Run the application
CMD ["./main"]
