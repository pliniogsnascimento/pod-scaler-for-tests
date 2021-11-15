FROM golang:1.17.2-alpine3.13
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .
ENTRYPOINT ["/app/main"]