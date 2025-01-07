FROM golang:1.23-alpine AS builder


WORKDIR /app


COPY go.mod go.sum ./


RUN go mod download


COPY . .


ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64


RUN go build -o myapp ./cmd


FROM alpine:latest


WORKDIR /app


COPY --from=builder /app/myapp /app/


EXPOSE 8080


CMD ["./myapp"]
