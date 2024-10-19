FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o emojify

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/emojify .

EXPOSE 8080

CMD ["./emojify"]
