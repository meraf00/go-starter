# Stage 1: Build
FROM golang:1.25-alpine AS builder

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Install build dependencies (needed for CGO)
RUN apk add --no-cache git

WORKDIR /src

# 1. Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# 2. Copy source and build
COPY . .
# Set CGO_ENABLED=0 for a static binary and use flags to strip debug info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/main ./cmd/api/main.go


# Stage 2: Final Runtime
FROM scratch

# Import users/groups from builder to scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy SSL certificates (essential to make HTTPS calls)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

# Copy the statically linked binary
COPY --from=builder /app/main .

# Include migration files for deployment
COPY ./store/migrations ./store/migrations

# Use the non-root user defined in the builder stage
USER appuser

EXPOSE 9090

ENTRYPOINT ["./main"]