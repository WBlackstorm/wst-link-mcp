# Stage 1: Build binary
FROM golang:alpine AS builder

WORKDIR /app

# Install ca-certificates (needed for TLS HTTP requests to LinkedIn)
RUN apk add --no-cache ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build lightweight static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /linkmcp ./cmd/linkmcp/main.go

# Stage 2: Minimal runtime image
FROM scratch

# Copy CA root certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder stage
COPY --from=builder /linkmcp /linkmcp

ENTRYPOINT ["/linkmcp"]
