# ==========================================
# STAGE 1: Build the Go Binary
# ==========================================
FROM golang:1.26-alpine AS builder

RUN apk update && apk upgrade && apk add --no-cache ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# We build the binary and name it 'todo-api'
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o todo-api ./cmd/api/main.go

# ==========================================
# STAGE 2: The Production Image
# ==========================================
FROM alpine:3.19.1

RUN apk update && apk upgrade && apk add --no-cache ca-certificates

# 1. Create a neutral directory for our app
WORKDIR /app

# 2. Create the non-root user FIRST
RUN adduser -D appuser

# 3. Copy the binary from Stage 1 into our neutral /app directory
COPY --from=builder /build/todo-api .

# 4. (Optional but good) Give appuser ownership of the binary
RUN chown appuser:appuser todo-api

# 5. Drop root privileges! We are now a restricted user.
USER appuser

EXPOSE 9090

# 6. Execute the binary
CMD ["./todo-api"]