# Build stage
FROM golang:1.25-alpine AS builder

ARG VERSION=dev

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w -X 'main.Version=${VERSION}'" -o btcleaner ./cmd/btcleaner

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/btcleaner .
COPY config.example.yaml /etc/btcleaner.yaml

# Create non-root user
RUN addgroup -S btcleaner && adduser -S btcleaner -G btcleaner
USER btcleaner

# Expose web UI port (for future use)
EXPOSE 8888

ENTRYPOINT ["/app/btcleaner"]
