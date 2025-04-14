# syntax=docker/dockerfile:1
FROM golang:1.24-bookworm AS builder

ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o mcpserver cmd/mcpserver/main.go

FROM gcr.io/distroless/static-debian11
COPY --from=builder /app/mcpserver .
ENTRYPOINT ["/mcpserver"]
