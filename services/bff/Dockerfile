# Build stage
FROM golang:1.25 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bff ./services/bff/cmd

# Final stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bff .

# Copy config
COPY services/bff/config/config.yml ./config.yml

EXPOSE 8080

CMD ["./bff"]
