# ---------- builder ----------
# Sesuaikan dengan requirement go.work (>= 1.24.5)
FROM golang:1.24.5-alpine AS builder

RUN apk add --no-cache git
WORKDIR /src

# Bawa workspace & go.mod minimal biar cache dependency kepake
COPY go.work ./
# COPY go.work.sum ./   # kalau ada, boleh ikut

COPY shared/go.mod shared/
COPY services/user-service/go.mod services/user-service/
# (opsional)
# COPY services/user-service/go.mod services/user-service/ || true

# Bawa seluruh source repo (karena replace ../../shared)
COPY . .

# Build biner user-service
WORKDIR /src/services/user-service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/user-service ./cmd

# Copy migrations ke output
# RUN mkdir -p /out/migrations && cp -r ./migrations/* /out/migrations/

# ---------- runtime ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates curl && adduser -D -H -s /sbin/nologin app
WORKDIR /app

COPY --from=builder /out/user-service /app/user-service
# COPY --from=builder /out/migrations /app/migrations

EXPOSE 9001
USER app
ENTRYPOINT ["./user-service"]
