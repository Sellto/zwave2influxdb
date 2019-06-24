FROM golang:1.12.4-alpine as dev
RUN apk add git
RUN go get -v github.com/clbanning/mxj
RUN go get -v github.com/influxdata/influxdb1-client/v2
ADD zwave2influxdb.go zwave2influxdb.go
RUN apk add build-base
RUN go build zwave2influxdb.go
CMD sh

FROM alpine
COPY --from=dev /go/zwave2influxdb .
CMD ./zwave2influxdb
