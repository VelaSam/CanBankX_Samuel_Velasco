FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /identity ./cmd/identity/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /accounts ./cmd/accounts/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /transactions ./cmd/transactions/main.go

FROM alpine:latest

WORKDIR /
COPY --from=builder /identity /identity
COPY --from=builder /accounts /accounts
COPY --from=builder /transactions /transactions

EXPOSE 8081 8082 8083
