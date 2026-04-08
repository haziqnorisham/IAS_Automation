# Use alpine Linux for small image size
FROM alpine:latest

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /app

# Copy the binary from your local machine
COPY --chmod=0755 ias_automation_v0.01 .

# Copy .env file
COPY .env .

# Expose port (adjust to your HTTP_SERVER_PORT)
EXPOSE 8080

# Run the binary
CMD ["/app/ias_automation_v0.01"]