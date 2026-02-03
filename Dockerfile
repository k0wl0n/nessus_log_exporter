# ── Stage 1: build ──────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /nessus_log_exporter .

# ── Stage 2: runtime ────────────────────────────────────────────────
FROM alpine:3.19
RUN apk add --no-cache ca-certificates

COPY --from=builder /nessus_log_exporter /usr/local/bin/nessus_log_exporter

EXPOSE 19835

ENTRYPOINT ["nessus_log_exporter"]
