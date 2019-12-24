FROM golang:latest as builder
ADD . /url-operator
RUN  cd /url-operator && go build
FROM alpine:latest as prod
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /url-operator/url-operator .
RUN chmod +x ./url-operator
CMD ["./url-operator"]