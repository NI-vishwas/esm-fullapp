# --- Stage 1: Build binary ---
FROM golang:1.25-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o event-service .

# --- Stage 2: Minimal Runtime Container ---
FROM alpine:3.19

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/event-service .

USER appuser

EXPOSE 8081

CMD ["./event-service"]