FROM golang:latest as builder
RUN git clone https://github.com/tx19980520/url-operator.git && git checkout master
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o url-operator .

FROM alpine:latest as prod
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder ~/url-operator/url-operator .
RUN chmod +x ./url-operator
CMD ["./url-operator"]