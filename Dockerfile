FROM golang:1.5.1

WORKDIR /go/src/github.com/marcusolsson/goddd
ADD . /go/src/github.com/marcusolsson/goddd

RUN go get github.com/tools/godep

RUN godep go install github.com/marcusolsson/goddd

ENTRYPOINT /go/bin/goddd

EXPOSE 8080

