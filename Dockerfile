FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ollama-dns-router

FROM alpine:3.21
RUN addgroup -S nonroot && \
    adduser -S nonroot -G nonroot
COPY --from=builder /app/ollama-dns-router /ollama-dns-router
USER nonroot
ENTRYPOINT ["/ollama-dns-router"]
