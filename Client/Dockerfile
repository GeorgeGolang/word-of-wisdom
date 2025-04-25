FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /wisdom-client ./main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /wisdom-client .
CMD ["./wisdom-client"]