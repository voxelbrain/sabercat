FROM golang:onbuild

USER 1

EXPOSE 8080

CMD app --address 0.0.0.0:8080
