# Build stage for Go backend
FROM golang:1.21 AS go-builder

WORKDIR /app

# Copy Go module files from mainote_server
COPY mainote_server/go.mod mainote_server/go.sum ./
RUN go mod download

# Copy Go source code from mainote_server
COPY mainote_server/cmd/ cmd/
COPY mainote_server/internal/ internal/

# Ensure modules are properly resolved
RUN go mod tidy

# Build Go backend
RUN CGO_ENABLED=0 GOOS=linux go build -o go-backend cmd/server/main.go

# Final stage with Python runtime
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libpq-dev \
    supervisor \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements and setup files first to leverage Docker cache
COPY requirements.txt setup.py README.md ./
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application
COPY . .

# Install the package
RUN pip install -e .

# Copy Go backend binary from builder stage
COPY --from=go-builder /app/go-backend ./go-backend

# Make startup script executable
RUN chmod +x start.sh

# Copy supervisord configuration
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV PORT=8080

# Run supervisord to manage both processes
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
