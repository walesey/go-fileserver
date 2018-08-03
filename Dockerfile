FROM golang:1.7.3

COPY . /go/src/github.com/walesey/go-fileserver

RUN go install github.com/walesey/go-fileserver

CMD ./go-fileserver -server -path ./serverFiles