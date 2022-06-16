FROM golang:latest as build

WORKDIR /usr/src/tinkoff-invest-contest

COPY go.mod ./
RUN go mod download && go mod verify

COPY . .
COPY ./.env .
RUN go build -o trade ./cmd/trade/main.go

CMD ./trade --mode sandbox
