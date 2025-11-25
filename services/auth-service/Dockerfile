# ---------- builder ----------
# Sesuaikan dengan requirement go.work (>= 1.24.5)
FROM golang:1.24.5-alpine AS builder

RUN apk add --no-cache git
WORKDIR /src

# Bawa workspace & go.mod minimal biar cache dependency kepake
COPY go.work ./
# COPY go.work.sum ./   # kalau ada, boleh ikut

COPY shared/go.mod shared/
COPY services/auth-service/go.mod services/auth-service/
# (opsional)
# COPY services/user-service/go.mod services/user-service/ || true

# Bawa seluruh source repo (karena replace ../../shared)
COPY . .

# Build biner auth-service
WORKDIR /src/services/auth-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/auth-service ./cmd

# Copy migrations ke output
RUN mkdir -p /out/migrations && cp -r ./migrations/* /out/migrations/

# ---------- runtime ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates curl tzdata && \
    adduser -D -H -s /sbin/nologin app

ENV TZ=Asia/Jakarta
RUN cp /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && \
    echo "Asia/Jakarta" > /etc/timezone
WORKDIR /app

COPY --from=builder /out/auth-service /app/auth-service
COPY --from=builder /out/migrations /app/migrations

EXPOSE 9001
USER app
ENTRYPOINT ["./auth-service"]
