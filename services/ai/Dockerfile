FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -trimpath -o /out/server ./cmd/server

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -D -g '' -u 10001 appuser

WORKDIR /app
COPY --from=builder /out/server /app/server

ENV PORT=8080
USER appuser
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
     CMD wget -qO- http://localhost:${PORT}/health >/dev/null 2>&1 || exit 1
#    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1

ENTRYPOINT ["/app/server"]
