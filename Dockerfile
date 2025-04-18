# syntax=docker/dockerfile:1
FROM golang:1.24-bookworm AS builder

ENV GOOS=linux

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" -o mcpserver cmd/mcpserver/main.go

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/mcpserver /mcpserver
USER nonroot:nonroot
ENTRYPOINT ["/mcpserver"]
