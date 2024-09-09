FROM golang:1.23-alpine3.20 AS build

WORKDIR /src

COPY . .

RUN go mod download 

RUN go build -o bin/main ./cmd/main/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /src/bin/main .

CMD ["./main"]