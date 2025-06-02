FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements and setup files first to leverage Docker cache
COPY requirements.txt setup.py README.md ./
RUN pip install --no-cache-dir -r requirements.txt

# Copy the rest of the application
COPY . .

# Install the package
RUN pip install -e .

# Make startup script executable
RUN chmod +x start.sh

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV PORT=8080

# Run the startup script
CMD ["./start.sh"]
