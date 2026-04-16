# STAGE 1: Build
# Updated to 1.26 to match your go.mod requirement
FROM golang:1.26-alpine AS builder

# Fixed apk command (removed invalid flag)
RUN apk update && apk upgrade && apk add --no-cache ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o api ./cmd/api/main.go

# STAGE 2: Final Image
FROM alpine:3.19.1

RUN apk update && apk upgrade && apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/api .
COPY .env .

RUN adduser -D appuser
USER appuser

EXPOSE 9090
CMD ["./api"]