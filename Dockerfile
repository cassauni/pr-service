FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pr-service ./cmd/main.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates


COPY --from=builder /src/pr-service /usr/local/bin/pr-service
RUN chmod +x /usr/local/bin/pr-service

WORKDIR /

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/pr-service"]