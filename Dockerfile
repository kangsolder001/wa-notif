# syntax=docker/dockerfile:1
# Image wa-notif untuk Raspberry Pi 4 (linux/arm64). Build silang tanpa emulasi:
#   docker buildx build --platform linux/arm64 -t wa-notif:<tag> --load .
# Tahap build berjalan di arsitektur mesin build; hanya biner statis yang disalin ke image akhir.

FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS build
ARG TARGETOS=linux
ARG TARGETARCH=arm64
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY cmd ./cmd
COPY config ./config
COPY internal ./internal
COPY pkg ./pkg
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/wa-notif ./cmd/server/main.go

# distroless static: tanpa shell/paket, pengguna nonroot = UID/GID 65532
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/wa-notif /usr/local/bin/wa-notif
USER 65532:65532
EXPOSE 3030
ENV CONFIG_PATH=/etc/wa-notif/config.json
ENTRYPOINT ["/usr/local/bin/wa-notif"]
