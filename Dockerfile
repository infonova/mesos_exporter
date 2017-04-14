FROM golang:1.8-alpine

EXPOSE 9110

RUN addgroup exporter \
 && adduser -S -G exporter exporter

COPY ./ /mesos_exporter
WORKDIR /mesos_exporter
RUN apk --update add --virtual build-deps git \
&& go get -d \
&& go build \
&& apk del --purge build-deps \
&& rm -rf /go/bin /go/pkg /var/cache/apk/*

USER exporter

ENTRYPOINT [ "/mesos_exporter/mesos_exporter" ]
