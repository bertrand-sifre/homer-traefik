# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23.4-alpine3.21 AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG TARGETPLATFORM
RUN case "${TARGETPLATFORM}" in \
      "linux/amd64") GOARCH=amd64 ;; \
      "linux/arm64") GOARCH=arm64 ;; \
      "linux/arm/v7") GOARCH=arm GOARM=7 ;; \
      *) echo "Unsupported platform: ${TARGETPLATFORM}" && exit 1 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} GOARM=${GOARM:-} go build -o homer-traefik

# Final stage with minimal image
FROM --platform=$TARGETPLATFORM alpine:3.21

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/homer-traefik .

# Create volume for Homer configuration
VOLUME /app/config

# Set working directory to configuration directory
WORKDIR /app/config

# Run the application
CMD ["/app/homer-traefik"]