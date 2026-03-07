FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /canbankx ./cmd/server/main.go

FROM alpine:latest

WORKDIR /
COPY --from=builder /canbankx /canbankx

EXPOSE 8080

ENTRYPOINT ["/canbankx"]
