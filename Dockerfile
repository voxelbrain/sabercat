FROM alpine

RUN apk update && apk add ca-certificates

COPY ./sabercat /usr/local/bin/sabercat

USER 1

EXPOSE 8080

CMD sabercat --address 0.0.0.0:8080
