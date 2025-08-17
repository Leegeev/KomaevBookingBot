# syntax=docker/dockerfile:1

########################
# Build stage
########################
FROM golang:1.22.3-alpine3.20 AS builder

WORKDIR /src
RUN apk update && apk add --no-cache git

# Сначала зависимости — чтобы кешировать
COPY go.mod go.sum ./
RUN go mod download

# Затем всё остальное
COPY . .

# Собираем бинарь
# (GOARCH оставляем по умолчанию, чтобы собирать под текущую платформу)
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" \
    -o /out/meetingbot ./cmd/MeetingBotV1

########################
# Runtime stage
########################
FROM alpine:3.20

# tzdata нужно, чтобы time.LoadLocation("Europe/Moscow") работал корректно
RUN apk update && apk add --no-cache ca-certificates tzdata && \
    adduser -D -H -s /sbin/nologin appuser

# ВАЖНО: работаем из /src — туда смотрит твой загрузчик config.yaml (через runtime.Caller)
WORKDIR /src

# Копируем бинарь
COPY --from=builder /out/meetingbot /src/meetingbot

# config.yaml и .env мы примонтируем в compose как тома (чтобы менять без пересборки)
USER appuser

ENTRYPOINT ["/src/meetingbot"]
