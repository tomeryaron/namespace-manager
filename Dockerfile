# Simple runtime image - no build step needed
FROM alpine:latest

# Install CA certificates for HTTPS (needed for Kubernetes API)
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy the pre-built binary from your local machine
COPY namespace-manager .

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./namespace-manager"]