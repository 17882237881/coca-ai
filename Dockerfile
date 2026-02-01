# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Set Proxy for China
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o coca-server ./cmd/server

# Final Stage
FROM alpine:latest

WORKDIR /app

# Install timezone data
RUN apk add --no-cache tzdata

COPY --from=builder /app/coca-server .

EXPOSE 8080

CMD ["./coca-server"]
