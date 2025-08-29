# --- Build Stage ---
FROM golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the application
# -ldflags="-w -s" reduces the binary size
# CGO_ENABLED=0 for static linking
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /webnote .

# --- Final Stage ---
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the build stage
COPY --from=builder /webnote /app/webnote

# Copy static files and templates
COPY index.html .
COPY static ./static

# Create the notes storage directory and set permissions
# Using a non-root user is a security best practice.
# The user ID 1001 is arbitrary but common for non-root users.
RUN mkdir -p notes && \
    chown -R 1001:1001 notes && \
    chmod -R 755 notes

# Create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
# Copy the entrypoint script
COPY entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT ["entrypoint.sh"]

USER appuser


# Expose the port
EXPOSE 8080

# Set the startup command
CMD ["/app/webnote"]