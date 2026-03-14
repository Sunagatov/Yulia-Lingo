# =============================================================================
# BUILD STAGE — Modern 2026 approach with BuildKit cache mounts
# =============================================================================
FROM golang:1.24-alpine AS builder

ARG VERSION=dev

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./

# Download dependencies with BuildKit cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy source code
COPY . .

# Build with cached dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o yulia-lingo ./cmd/app

# =============================================================================
# RUNTIME STAGE
# =============================================================================
FROM scratch

ARG VERSION=dev

LABEL maintainer="Yulia-Lingo Team" \
      version="${VERSION}" \
      description="Yulia-Lingo Telegram Bot"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/yulia-lingo /yulia-lingo
COPY --from=builder /app/resource /resource

USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/yulia-lingo"]
