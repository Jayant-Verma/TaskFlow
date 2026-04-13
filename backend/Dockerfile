FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api_server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/api_server .
COPY migrations ./migrations
EXPOSE 8080
CMD ["./api_server"]