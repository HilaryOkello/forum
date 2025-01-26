# Start from the official Go image
FROM golang:1.23.4-alpine

# Set working directory inside the container
WORKDIR /app

# Install git and gcc (for sqlite)
RUN apk add --no-cache git gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN go build -o forum-app

# Expose port 8080
EXPOSE 8080

# Run the executable
CMD ["./forum-app"]