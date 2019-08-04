FROM golang:1.12-buster AS build


WORKDIR /build

COPY . .

RUN make build

