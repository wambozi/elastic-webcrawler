FROM golang:1.13.5-alpine3.11

ARG ENV_ID

RUN apk --update add bash wget dpkg-dev

RUN addgroup -S elastic && adduser -S elastic -G elastic

COPY --chown=elastic:elastic ./bin/elastic-webcrawler /opt/bin/elastic-webcrawler
COPY --chown=elastic:elastic ./conf /conf

USER elastic

WORKDIR /opt/bin

CMD [ "./elastic-webcrawler" ]
