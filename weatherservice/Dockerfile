FROM golang:1.11.0 as testing

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...

# Run the test
RUN go test 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .


FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/app/app .
CMD ["./app"]  