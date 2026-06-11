# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Git zaroori hota hai dependencies download karne ke liye
RUN apk add --no-cache git

WORKDIR /app

# Sabhi files copy karein
COPY . .

# IMPORTANT: Ye commands missing dependencies ko theek karengi
RUN go mod tidy
RUN go mod download

# Binary build karein
RUN go build -o main .

# Stage 2: Minimal image for running the app
FROM alpine:latest
# Certificates install karein taaki MongoDB Atlas (SSL) se connect ho sake
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Builder stage se binary copy karein
COPY --from=builder /app/main .

# Port expose karein
EXPOSE 8080

# App run karein
CMD ["./main"]