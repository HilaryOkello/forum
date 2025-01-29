# Build stage
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application with additional optimizations
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o forum-app

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the executable and ALL project files from builder stage
COPY --from=builder /app/forum-app .
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/db ./db
COPY --from=builder /app/forum.db .


EXPOSE 8080

CMD ["./forum-app"]