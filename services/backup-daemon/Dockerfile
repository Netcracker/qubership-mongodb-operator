FROM ghcr.io/netcracker/qubership-backup-daemon-go:test_branch

ENV MONGO_HOME=/opt/mongodb-backup

RUN apk update && apk add --no-cache mongodb-tools jq zip unzip bash

RUN echo "https://dl-cdn.alpinelinux.org/alpine/edge/main" > /etc/apk/repositories  \
    && echo "https://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories \
    && apk add --update --upgrade --no-cache \
        libcurl curl

ENV DATABASES_FLAG="-d" \
    DBMAP_FLAG="-m"

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
