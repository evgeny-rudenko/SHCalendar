# Multi-stage build for the calendar service
# 1) Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Enable static build (CGO not required for modernc.org/sqlite)
ENV CGO_ENABLED=0 GOOS=linux

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN go build -ldflags="-s -w" -o /out/server ./

# 2) Runtime stage (distroless, non-root)
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Optional: directory for persistent data (mounted via Compose)
# The app will use DB_PATH=/data/calendar.db

COPY --from=builder /out/server /server

EXPOSE 8085

ENV PORT=8085

ENTRYPOINT ["/server"]
