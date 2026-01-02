# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o calsun .

# Runtime stage
FROM alpine:latest

# Install timezone data for proper timezone handling
RUN apk add --no-cache tzdata

WORKDIR /app

# Copy the binary
COPY --from=builder /app/calsun .

# Expose port
EXPOSE 8080

# Run the server
CMD ["./calsun"]
