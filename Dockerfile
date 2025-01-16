# Etapa 1: Build
FROM golang:1.23.3

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo -o main ./cmd/main.go

EXPOSE 8080

CMD ["./main"]