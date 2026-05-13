FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o kuda ./cmd/server/main.go

# give the file execution permissions
RUN chmod +x ./kuda

FROM alpine:latest
WORKDIR /app
# Install ca-certificates in case workers need to call HTTPS webhooks
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/kuda .
EXPOSE 8000
CMD ["./kuda"]