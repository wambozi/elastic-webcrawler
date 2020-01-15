FROM golang:1.13.5-alpine3.11

ARG ENV_ID

RUN apk --update add bash wget dpkg-dev

RUN addgroup -S elastic && adduser -S elastic -G elastic

COPY --chown=elastic:elastic ./bin/elastic-webcrawler /opt/bin/elastic-webcrawler
<<<<<<< Updated upstream
COPY --chown=elastic:elastic ./conf /conf
=======
COPY --chown=elastic:elastic ./conf /opt/bin/conf

RUN chmod -R 755 /opt/bin/conf
>>>>>>> Stashed changes

USER elastic

WORKDIR /opt/bin

CMD [ "./elastic-webcrawler" ]
