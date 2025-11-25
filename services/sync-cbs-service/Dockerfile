# ---------- builder ----------
# Sesuaikan dengan requirement go.work (>= 1.24.5)
FROM golang:1.24.5-alpine AS builder

RUN apk add --no-cache git
WORKDIR /src

# Bawa workspace & go.mod minimal biar cache dependency kepake
COPY go.work ./
# COPY go.work.sum ./   # kalau ada, boleh ikut

COPY shared/go.mod shared/
COPY services/sync-cbs-service/go.mod services/sync-cbs-service/
# (opsional)
# COPY services/sync-cbs-service/go.mod services/sync-cbs-service/ || true

# Bawa seluruh source repo (karena replace ../../shared)
COPY . .

# Build biner sync-cbs-service
WORKDIR /src/services/sync-cbs-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sync-cbs-service ./cmd

# Copy migrations ke output
# RUN mkdir -p /out/migrations && cp -r ./migrations/* /out/migrations/

# ---------- runtime ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates curl tzdata && \
    adduser -D -H -s /sbin/nologin app

ENV TZ=Asia/Jakarta
RUN cp /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && \
    echo "Asia/Jakarta" > /etc/timezone
WORKDIR /app

COPY --from=builder /out/sync-cbs-service /app/sync-cbs-service
# COPY --from=builder /out/migrations /app/migrations

EXPOSE 9001
USER app
ENTRYPOINT ["./sync-cbs-service"]
