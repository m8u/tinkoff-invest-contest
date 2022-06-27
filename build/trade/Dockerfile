FROM golang:1.18.3

WORKDIR /usr/src/tinkoff-invest-contest

COPY go.mod ./
RUN go mod download && go mod verify

COPY . .
COPY .env .
RUN go build -o trade ./cmd/trade/main.go

STOPSIGNAL SIGINT
CMD ["./trade", "--mode",  "sandbox"]
