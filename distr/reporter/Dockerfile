FROM golang:1.19.2-alpine3.16

WORKDIR /app

COPY ../../ /app

RUN go mod download
RUN go mod tidy
RUN go build  -o /app/bin/report /app/cmd/report