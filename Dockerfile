# Build stage
FROM golang:1.23 AS builder

# Set working directory
WORKDIR /app

# Install build dependencies

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o main .


# Final stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy database directory
COPY --from=builder /app/database ./database

# Create data directory for SQLite database
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release
ENV DB_PATH=/app/data/app.db

# Run the application
CMD ["./main"]