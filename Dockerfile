FROM alpine:latest

COPY ./log.alot /usr/bin/alot

CMD ["alot"]
