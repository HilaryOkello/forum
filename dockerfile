# Start from the official Go image
FROM golang:1.23.4-alpine as builder

# Set working directory inside the container
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod and sum files first to leverage Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy the source code
COPY . .

# Build the application statically to minimize dependencies
RUN go build -o forum-app

# Use a smaller image for the final container
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy only the executable from the builder stage
COPY --from=builder /app/forum-app .

# Expose port 8080
EXPOSE 8080

# Run the executable
CMD ["./forum-app"]

