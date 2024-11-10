FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY cmd ./cmd
COPY internal ./internal
COPY go.mod go.sum ./

RUN export GOPROXY=direct
RUN go mod download

WORKDIR /build/cmd/auth

RUN go build -o /build/auth .

FROM alpine:latest
WORKDIR /app
COPY config/local.yaml ./config/
# Копируем бинарный файл из предыдущего этапа
COPY --from=builder /build/auth /app/auth
ENV CONFIG_PATH=./config/local.yaml
CMD ["/app/auth"]
