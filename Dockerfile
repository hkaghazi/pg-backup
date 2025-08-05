FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o pg-backup .

FROM alpine:latest

RUN apk add --no-cache postgresql-client ca-certificates

WORKDIR /app

COPY --from=builder /app/pg-backup .
COPY config.yaml .

CMD ["./pg-backup"]
