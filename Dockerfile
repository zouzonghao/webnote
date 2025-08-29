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
COPY index.html history.html .
COPY static ./static

# Use minified assets in production
RUN rm /app/static/style.css /app/static/script.js && \
    mv /app/static/style.min.css /app/static/style.css && \
    mv /app/static/script.min.js /app/static/script.js

# Create the notes storage directory
RUN mkdir -p notes


# Expose the port
EXPOSE 8080

# Set the startup command
CMD ["/app/webnote"]