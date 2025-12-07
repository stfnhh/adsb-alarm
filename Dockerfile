# checkov:skip=CKV_DOCKER_2: Healthchecks are handled by Kubernetes, not Docker

FROM alpine:3.21 AS certs
RUN apk add --no-cache ca-certificates

FROM golang:1.25.3-alpine3.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal ./internal
RUN apk add --no-cache upx && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o adsb-monitor . && \
    upx --best --lzma adsb-monitor

FROM scratch
USER 65532:65532
WORKDIR /app
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/adsb-monitor .
HEALTHCHECK NONE
ENTRYPOINT ["/app/adsb-monitor"]
