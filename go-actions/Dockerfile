# Use official Go image
FROM golang:1.22

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the rest of your application
COPY . .

# Build the app
RUN go build -o main .

# Expose port
EXPOSE 8002

# Run the binary
CMD ["./cmd/api/main"]
