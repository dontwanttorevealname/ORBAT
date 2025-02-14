
# Build stage
FROM golang:1.22-alpine AS builder

# Install git and build tools
RUN apk add --no-cache git build-base

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates and tzdata
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user
RUN adduser -D appuser

# Set working directory
WORKDIR /app

# Copy binary and templates from builder
#COPY --from=builder /app/main .
#COPY --from=builder /app/templates ./templates/
COPY --from=builder /app .

# Verify templates directory exists and has content
RUN ls -la /app/templates/

# Set ownership
RUN chown -R appuser:appuser /app

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Command to run with explicit health check
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

# Command to run
CMD ["./main"]

