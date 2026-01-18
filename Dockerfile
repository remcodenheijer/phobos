# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate templ files and build
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy static assets
COPY --from=builder /app/static ./static
COPY --from=builder /app/assets ./assets

# Expose port
EXPOSE 3000

# Run the server
CMD ["./server", "-port", "3000"]
