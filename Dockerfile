# ===== Build stage =====
FROM golang:1.23-bookworm AS builder

# Install C compiler for CGO (required by go-sqlite3)
RUN apt-get update && apt-get install -y gcc libc6-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /build/Backend

# Copy Go module files first for Docker layer caching
COPY Backend/go.mod Backend/go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

# Copy full Backend source and build
COPY Backend/ ./
ENV CGO_ENABLED=1
ENV GOTOOLCHAIN=auto
RUN go build -o /build/server ./cmd/server

# ===== Runtime stage =====
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates libc6 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy compiled binary
COPY --from=builder /build/server ./server

# Copy frontend static files
COPY *.html ./static/
COPY *.css ./static/
COPY *.js ./static/
COPY *.png ./static/
COPY *.txt ./static/
COPY *.webmanifest ./static/
COPY ["Nessie Audio 2026.jpg", "./static/Nessie Audio 2026.jpg"]
COPY ["Product Photos", "./static/Product Photos"]
COPY Music ./static/Music

# Create runtime directories
RUN mkdir -p /app/logs /app/backups

EXPOSE 8080

# Database will be created by migrations on first boot
CMD ["./server"]
