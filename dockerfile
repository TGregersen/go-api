# syntax=docker/dockerfile:1
FROM golang:1.23.4

WORKDIR $GOPATH/src

COPY . .
RUN go mod download
RUN go build -o /receipts

EXPOSE 9000

CMD ["/receipts"]