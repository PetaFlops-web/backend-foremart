# Build
FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/web/main.go

# Runner
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata wget

COPY --from=builder /app/server .

ENV PORT=8080
EXPOSE ${PORT}

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/swagger/index.html || exit 1

CMD ["./server"]
