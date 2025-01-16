# Build stage
FROM golang:1.23.4-alpine3.21 AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o homer-traefik

# Final stage with minimal image
FROM alpine:3.21

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/homer-traefik .

# Create volume for Homer configuration
VOLUME /app/config

# Set working directory to configuration directory
WORKDIR /app/config

# Run the application
CMD ["/app/homer-traefik"]