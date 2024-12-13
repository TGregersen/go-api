# syntax=docker/dockerfile:1
FROM golang:1.23.4

WORKDIR $GOPATH/app

RUN go build -o ~/GIT/receipts.go

COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080

CMD ["~/GIT/receipts"]