FROM golang:1.25

WORKDIR /usr/src/app

COPY go.* ./
RUN go mod download

COPY . .

RUN go build -v -o main main.go

EXPOSE 8080

CMD ["./main"]
