FROM golang

WORKDIR /go/src/gofd
COPY . .

RUN go get -d  -v ./...
RUN go build -o gofd


CMD ["./gofd"]