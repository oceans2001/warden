FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o warden-api ./cmd/warden-api

FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl

COPY --from=builder /app/warden-api /usr/local/bin/warden-api

RUN curl -L --fail https://github.com/cilium/hubble/releases/latest/download/hubble-linux-arm64.tar.gz -o /tmp/hubble.tar.gz \
    && tar -xzf /tmp/hubble.tar.gz -C /usr/local/bin \
    && rm /tmp/hubble.tar.gz

ENTRYPOINT ["/usr/local/bin/warden-api"]