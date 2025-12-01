FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o worker ./cmd/worker

FROM ubuntu:22.04

# pdfcpu + ImageMagick
RUN apt update && apt install -y \
    imagemagick \
    ghostscript \
    poppler-utils \
    && rm -rf /var/lib/apt/lists/*

# Install pdfcpu CLI
RUN curl -L https://github.com/pdfcpu/pdfcpu/releases/download/v0.9.0/pdfcpu_0.9.0_Linux_x86_64.tar.xz \
    -o pdfcpu.tar.xz && \
    tar -xf pdfcpu.tar.xz && \
    mv pdfcpu /usr/local/bin/pdfcpu && \
    rm -rf pdfcpu.tar.xz

WORKDIR /app
COPY --from=builder /app/worker .
COPY .env /app/.env

CMD ["./worker"]
