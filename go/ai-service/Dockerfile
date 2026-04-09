FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ai-service ./cmd/server

FROM alpine:3.19

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 1001 appuser

COPY --from=builder /ai-service /ai-service

USER appuser

EXPOSE 8093
ENTRYPOINT ["/ai-service"]
