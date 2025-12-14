# checkov:skip=CKV_DOCKER_2: Healthchecks are handled by Kubernetes, not Docker

FROM --platform=$BUILDPLATFORM golang:1.25.5-alpine3.23 AS builder
WORKDIR /app
RUN apk add --no-cache ca-certificates upx

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o adsb-monitor . && \
    upx --best --lzma adsb-monitor

FROM scratch
USER 65532:65532
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/adsb-monitor .

HEALTHCHECK NONE
ENTRYPOINT ["/app/adsb-monitor"]
