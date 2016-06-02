FROM golang

COPY . $GOPATH/src/github.com/campact/sabercat/

RUN go get -d -v github.com/campact/sabercat/cmd/sabercat && \
  go install github.com/campact/sabercat/cmd/sabercat

USER 1

EXPOSE 8080

CMD sabercat --address 0.0.0.0:8080
