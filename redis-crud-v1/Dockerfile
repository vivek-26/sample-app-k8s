FROM golang:1.15.6-alpine3.12

RUN mkdir /redis-crud-v1

COPY . /redis-crud-v1

WORKDIR /redis-crud-v1

RUN go mod download

RUN go build -o application .