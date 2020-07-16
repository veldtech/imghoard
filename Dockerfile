FROM golang:1.14

EXPOSE 8080

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build -o app -v ./cmd/api

CMD ["./app"]