# Stage 1: Build
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o hotwheels-bot ./main.go

# Stage 2: Minimal Runtime
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S bot && adduser -S bot -G bot \
    && mkdir -p /var/data && chown bot:bot /var/data

COPY --from=builder /app/hotwheels-bot /hotwheels-bot

USER bot

VOLUME ["/var/data"]

EXPOSE 8080

ENTRYPOINT ["/hotwheels-bot"]
