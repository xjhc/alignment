# ---- Build Stage ----
# Use the official Go image to build our application
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go server binary for a Linux environment
# The output binary will be named "server"
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /server ./server/cmd/server/

# ---- Final Stage ----
# Use a minimal, non-root image for the final container
FROM gcr.io/distroless/static-debian11

# Copy the server binary from the builder stage
COPY --from=builder /server /server

# Expose the port the server listens on
EXPOSE 8080

# The command to run when the container starts
CMD ["/server"]