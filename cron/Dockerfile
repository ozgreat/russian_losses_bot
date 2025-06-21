FROM alpine:latest

RUN apk add --no-cache curl busybox-suid

COPY crontab /etc/crontabs/root

CMD ["crond", "-f"]
