FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN test -z "$(gofmt -l .)" && \
    go vet ./... && \
    go test ./... && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api-gateway-lite ./cmd/api-gateway-lite && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/bench-target ./cmd/bench-target && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/benchmark ./cmd/benchmark

FROM alpine:3.21

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && \
    adduser -S -G app app
COPY --from=build /out/api-gateway-lite /usr/local/bin/api-gateway-lite
COPY --from=build /out/bench-target /usr/local/bin/bench-target
COPY --from=build /out/benchmark /usr/local/bin/benchmark

USER app
EXPOSE 8080
ENTRYPOINT []
CMD ["api-gateway-lite"]
