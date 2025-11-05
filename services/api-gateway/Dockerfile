# Build
FROM golang:1.24.5-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.work ./
COPY shared/go.mod shared/
COPY services/api-gateway/go.mod services/api-gateway/
COPY . .
WORKDIR /src/services/api-gateway
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api-gateway ./cmd

# Runtime
FROM alpine:3.20
WORKDIR /app
RUN adduser -S -D -H app
COPY --from=builder /out/api-gateway /app/api-gateway
USER app
EXPOSE 8008
ENTRYPOINT ["./api-gateway"]
