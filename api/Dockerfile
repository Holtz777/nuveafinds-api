# ---- build stage ----
FROM golang:1.23-alpine AS build
WORKDIR /src

# Cache de deps: copia só go.mod (go.sum ainda não existe enquanto não temos deps externas)
COPY go.mod ./
RUN go mod download || true

COPY . .

# Binário estático, sem CGO, para rodar em scratch.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- runtime stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/server /app/server

ENV PORT=8080
EXPOSE 8080

# Healthcheck via /health
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:${PORT}/health || exit 1

ENTRYPOINT ["/app/server"]
