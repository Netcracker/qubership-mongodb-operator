FROM alpine:3.19.1

ENV WORKDIR=/opt/operator/

ENV GOSUMDB=off GOPRIVATE=github.com/Netcracker
RUN --mount=type=secret,id=GH_ACCESS_TOKEN git config --global url."https://$(cat /run/secrets/GH_ACCESS_TOKEN)@github.com/".insteadOf "https://github.com/"

ENV OPERATOR=/usr/local/bin/mongo-operator \
    USER_UID=1001 \
    USER_NAME=mongo-operator

RUN echo 'https://dl-cdn.alpinelinux.org/alpine/v3.19/main/' > /etc/apk/repositories \
    && apk add --no-cache openssl curl

COPY bin/mongo-operator ${OPERATOR}
COPY build/bin /usr/local/bin

RUN chmod +x /usr/local/bin/entrypoint
RUN  chmod +x /usr/local/bin/user_setup && /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
