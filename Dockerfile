# syntax=docker/dockerfile:1

# Stage 1: Build binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Enable auto toolchain download for Go 1.26+ requirement
ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compile static binary without CGO (Pure Go SQLite via modernc.org/sqlite)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/wa-notif ./cmd/server/main.go

# Stage 2: Minimal production runtime image
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy compiled binary and example config
COPY --from=builder /app/bin/wa-notif /app/wa-notif
COPY --from=builder /app/config.example.json /app/config.example.json

# Default storage directory for session persistence
RUN mkdir -p /app/storage

EXPOSE 3030

ENTRYPOINT ["/app/wa-notif"]
