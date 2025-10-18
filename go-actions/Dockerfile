# Use Go 1.24
FROM golang:1.24

# Set working directory
WORKDIR /app

# Set Go proxy for reliable module fetch
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy
RUN go mod download

# Copy the rest of your code
COPY . .

# Build the app from the correct main file path
RUN go build -o main ./cmd/api

# Expose the port your Go server listens on
EXPOSE 8002

# Run the compiled binary
CMD ["./main"]
