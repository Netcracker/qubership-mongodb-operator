FROM ghcr.io/netcracker/qubership-backup-daemon-go:0.1.27@sha256:2dbf4cac27195f13797c9049f803f9f6d4cb977ca3f9390d845f3c69f9d712b4

ENV MONGO_HOME=/opt/mongodb-backup

RUN apk update && apk add --no-cache mongodb-tools jq zip unzip bash

RUN echo "https://dl-cdn.alpinelinux.org/alpine/edge/main" > /etc/apk/repositories  \
    && echo "https://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories \
    && apk add --update --upgrade --no-cache \
        libcurl curl jq

ADD backup-daemon.conf /etc/backup-daemon.conf
ADD backup/ ${MONGO_HOME}/
ADD files/  ${MONGO_HOME}/
RUN chmod +x ${MONGO_HOME}/run.sh
RUN chmod +x ${MONGO_HOME}/scripts.sh

VOLUME /backup-storage
VOLUME /external
VOLUME /var/lib/yum/history
VOLUME /var/lib/machines
VOLUME /var/cache/ldconfig
VOLUME /root/.pki/nssdb
VOLUME /run/cryptsetup
VOLUME /etc
VOLUME /opt/backup
VOLUME /tmp

CMD ["sh", "-c", "$MONGO_HOME/run.sh"]
