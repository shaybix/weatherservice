FROM golang:1.11.0 as builder 
WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .


FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/app/app .
EXPOSE 8888
CMD ["./app"]  