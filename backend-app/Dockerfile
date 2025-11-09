# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server .

# Stage 2: Run
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8081
CMD ["/app/server"]