FROM golang:1.24-alpine AS builder

WORKDIR /app
    
RUN apk add --no-cache git
    
COPY go.mod go.sum ./

RUN go mod download
    
COPY . .
    
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd
    
# ---------------------------------------------------------------------------
    
FROM alpine:latest
    
WORKDIR /root/
    
COPY --from=builder /app/main .

COPY .env .env

COPY internal/migrations/scripts ./migrations

COPY frontend ./frontend

EXPOSE 8080
    
CMD ["./main"]