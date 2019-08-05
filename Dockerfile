#docker build -t jlarriba/logretriever:0.2 .
FROM golang:1.7-alpine
RUN mkdir -p /main
WORKDIR /main
ADD . /main
RUN go build ./main.go
CMD ["./main"]
