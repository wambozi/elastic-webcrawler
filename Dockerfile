FROM golang:1.13.5-alpine3.11

RUN apk --update add bash

RUN addgroup -S elastic && adduser -S elastic -G elastic

COPY --chown=elastic:elastic ./bin/elastic-webcrawler /opt/bin/elastic-webcrawler
COPY --chown=elastic:elastic ./conf /opt/bin/conf

RUN chmod -R 755 /opt/bin/conf/*

USER elastic

WORKDIR /opt/bin

CMD [ "./elastic-webcrawler" ]
