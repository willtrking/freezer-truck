FROM golang:1.14.4-alpine

RUN apk add build-base

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build -o main .

CMD ["/app/main"]