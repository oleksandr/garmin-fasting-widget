# syntax=docker/dockerfile:1

# Stage 1: Build the Go binary
FROM golang:alpine AS builder

WORKDIR /src

# Cache Go modules
COPY server/go.mod server/go.sum ./
RUN go mod download

# Build static binary without CGO
COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/fasting-server .

# Stage 2: Minimal runtime container
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup \
    && mkdir -p /data && chown -R appuser:appgroup /data

WORKDIR /app
COPY --from=builder /bin/fasting-server /app/fasting-server

USER appuser

EXPOSE 8080
VOLUME ["/data"]

ENV FASTING_PORT=8080
ENV FASTING_DB_PATH=/data/fasting.db

ENTRYPOINT ["/app/fasting-server"]
