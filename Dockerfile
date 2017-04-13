FROM golang:1.8-alpine

EXPOSE 9110

RUN addgroup exporter \
 && adduser -S -G exporter exporter

COPY ./ /mesos_exporter
WORKDIR /mesos_exporter
RUN apk --no-cache add --update git && go get -d && go build && apk del git pcre expat libcurl libssh2

USER exporter

ENTRYPOINT [ "/mesos_exporter/mesos_exporter" ]
