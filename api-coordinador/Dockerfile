FROM docker.io/library/golang:1.22-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download || true

COPY . .

RUN go build -o main .

CMD ["./main"]
