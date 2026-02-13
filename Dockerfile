FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/rules-engine-demo

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary
COPY --from=builder /app/server .

# Copy static files
COPY --from=builder /app/cmd/rules-engine-demo/static ./cmd/rules-engine-demo/static
COPY --from=builder /app/cmd/rules-engine-demo/templates ./cmd/rules-engine-demo/templates

# Copy examples
COPY --from=builder /app/examples ./examples

EXPOSE 8080

CMD ["./server"]
