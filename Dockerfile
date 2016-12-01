FROM golang:1.7.3

ADD . /go/src/github.com/walesey/go-fileserver
WORKDIR /go/src/github.com/walesey/go-fileserver

RUN go build

CMD ./go-fileserver server 3000