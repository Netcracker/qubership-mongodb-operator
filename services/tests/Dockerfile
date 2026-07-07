FROM ghcr.io/netcracker/qubership-docker-integration-tests:0.5.6
ENV ROBOT_OUTPUT=${ROBOT_HOME}/output \
    DISTR_DIR=/tmp/deps

RUN mkdir -p ${ROBOT_HOME} \
    && mkdir -p ${ROBOT_OUTPUT}

COPY pip.conf /etc/pip.conf
COPY requirements.txt ${ROBOT_HOME}/requirements.txt
COPY robot ${ROBOT_HOME}

RUN set -x \
    # && echo 'http://yumsrv03cn.netcracker.com/apk/alpine3.14/main/' > /etc/apk/repositories \
    && apk add --update --no-cache build-base 

# Upgrade all tools to avoid vulnerabilities

RUN set -x && apk upgrade --no-cache --available

RUN set -x \
    && pip3 install --upgrade pip setuptools \
    && pip3 install -r ${ROBOT_HOME}/requirements.txt \
    && rm -rf /var/cache/apk/*


USER root
RUN chmod -R 777 ${ROBOT_HOME}

# Expose the port
EXPOSE 8080
VOLUME ["${ROBOT_HOME}"]
