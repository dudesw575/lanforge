# Build stage
FROM golang:1.26 AS builder

WORKDIR /app

# Install git for cloning template repositories
# RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/main .
COPY ./templates/ ./templates/
COPY .env .

EXPOSE 8080

CMD ["./main"]