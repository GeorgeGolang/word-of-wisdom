FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /wisdom-server ./main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /wisdom-server .
EXPOSE 8080
CMD ["./wisdom-server"]