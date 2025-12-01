FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o worker ./cmd/worker

FROM ubuntu:22.04

# Image processing tools
RUN apt update && apt install -y \
    imagemagick \
    poppler-utils \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/worker .
COPY .env /app/.env

CMD ["./worker"]
