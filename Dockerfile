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

# Create the notes storage directory
RUN mkdir -p notes


# Expose the port
EXPOSE 8080

# Set the startup command
CMD ["/app/webnote"]