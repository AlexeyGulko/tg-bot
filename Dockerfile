FROM golang:1.19.2-alpine3.16

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY ./cmd/bot /app/cmd/bot
COPY ./internal /app/internal

RUN go mod download


RUN go build  -o /app/bin/bot /app/cmd/bot