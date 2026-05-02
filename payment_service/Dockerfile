FROM golang:1.25-alpine

WORKDIR /app

RUN apk add --no-cache git

ENV GOPROXY=direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/payment_service/main.go

CMD ["./main"]