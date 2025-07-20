# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies (handle case where go.sum doesn't exist)
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Production stage
FROM alpine:latest

# Install ca-certificates
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary and assets
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/archive ./archive
COPY --from=builder /app/public ./public

# Expose port
EXPOSE 8082

# Run the application
CMD ["./main"]
