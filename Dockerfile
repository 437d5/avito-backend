FROM golang:1.23-alpine3.20

WORKDIR /src

COPY . .

RUN go mod download 

RUN go build -o bin/main ./cmd/main/main.go

EXPOSE 8080

CMD ["/src/bin/main"]