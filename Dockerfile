FROM golang:1.18.4-alpine as base
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .

FROM base AS final
ENTRYPOINT ["/app/main"]