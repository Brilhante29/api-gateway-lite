FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY . .
RUN go mod tidy && \
    go test -v ./... && \
    CGO_ENABLED=0 go build -o /build/api-gateway-lite ./cmd/api-gateway-lite && \
    CGO_ENABLED=0 go build -o /build/bench-target ./cmd/bench-target && \
    CGO_ENABLED=0 go build -o /build/benchmark ./cmd/benchmark

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/api-gateway-lite /usr/local/bin/api-gateway-lite
COPY --from=builder /build/bench-target /usr/local/bin/bench-target
COPY --from=builder /build/benchmark /usr/local/bin/benchmark
EXPOSE 8080
CMD ["api-gateway-lite"]
