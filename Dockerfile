FROM golang:1.14

WORKDIR /go/src/app
ADD . /go/src/app/

RUN go mod download

RUN go build -o bin/statistics cmd/statistics/main.go

ENTRYPOINT ["bin/statistics"]

EXPOSE 8080
