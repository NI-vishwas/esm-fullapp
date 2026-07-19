# --- Stage 1: Build binary ---
FROM golang:1.25-alpine AS builder

# Install system dependencies (like git or certificates if needed)
RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /app

# Leverage Docker layer caching by copying module files first
COPY go.mod go.sum* ./
RUN go mod download

# Copy the rest of the microservice source code
COPY . .

# Build a statically linked, highly optimized Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o booking-service .

# --- Stage 2: Minimal Runtime Container ---
FROM alpine:3.19

# Security best practice: Add non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy CA Certificates from build stage (essential for S3/HTTPS requests)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /app/booking-service .

# Run under non-root user privileges
USER appuser

EXPOSE 8082

CMD ["./booking-service"]