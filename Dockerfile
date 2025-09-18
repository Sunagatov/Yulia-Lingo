FROM golang:1.21.3-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bot ./cmd/app

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/bot .
COPY --from=builder /app/resource ./resource
COPY --from=builder /app/.env .

CMD ["./bot"]