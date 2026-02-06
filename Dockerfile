# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary with hardening flags
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o calsun .

# Runtime stage — minimal image, no shell or package manager
FROM scratch

# Import timezone data and CA certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/calsun /calsun

# Run as non-root user (UID 65534 = nobody)
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/calsun"]
