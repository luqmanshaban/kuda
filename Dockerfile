FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o kuda .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/kuda .
EXPOSE 8000
CMD ["./kuda"]