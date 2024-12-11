FROM golang:1.19.4

WORKDIR $GOPATH/src

COPY . .

RUN yarn install --production

RUN go mod download
RUN go build -o ~/GIT

EXPOSE 8080

CMD ["~/GIT"]