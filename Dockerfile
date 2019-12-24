FROM golang:latest as builder
ADD . /url-operator
RUN  cd /url-operator && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
FROM alpine:latest as prod
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /url-operator/app .
RUN chmod +x ./app
CMD ["./app"]