FROM golang:1.24

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8000

# Run the application
CMD ["go", "run", "./cmd/server/main.go"]
