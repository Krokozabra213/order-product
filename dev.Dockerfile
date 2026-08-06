
FROM golang:1.25-alpine AS builder
RUN echo "https://mirror.yandex.ru/mirrors/alpine/v3.18/main" > /etc/apk/repositories
RUN echo "https://mirror.yandex.ru/mirrors/alpine/v3.18/community" >> /etc/apk/repositories
RUN apk add --no-cache git bash curl
WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 go build -o order-product ./cmd/order-product

# Финальный образ
FROM alpine:3.18
RUN echo "https://mirror.yandex.ru/mirrors/alpine/v3.18/main" > /etc/apk/repositories
RUN echo "https://mirror.yandex.ru/mirrors/alpine/v3.18/community" >> /etc/apk/repositories
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/order-product .
COPY --from=builder /app/configs ./configs

RUN chmod +x /app/order-product
CMD ["./order-product"]
