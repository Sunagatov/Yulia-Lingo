FROM golang:1.21.3 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -a -installsuffix cgo -o yulia-lingo-backend

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/yulia-lingo-backend .
EXPOSE 8443
CMD ["./yulia-lingo-backend"]
