FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init -g cmd/server/main.go
RUN go build -o bin/server ./cmd/server

FROM alpine:latest

RUN apk update && apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/bin/server .

EXPOSE 8080

CMD ["./server"]
