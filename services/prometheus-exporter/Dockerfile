# Use the latest official Alpine image
FROM alpine:latest

# Define the TARGETARCH argument (which will be automatically set by Docker during the build)
ARG TARGETARCH

# Update the APK repositories to use the official Alpine repository
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/edge/main' > /etc/apk/repositories \
    && apk add --no-cache openssl ca-certificates curl jq bash 

RUN echo "https://dl-cdn.alpinelinux.org/alpine/edge/main" > /etc/apk/repositories  \
    && echo "https://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories \
    && apk add --update --upgrade --no-cache \
        openssl ca-certificates curl jq bash libssl3 libcrypto3

# Download the latest MongoDB Exporter binary based on the architecture (amd64 or arm64)
RUN wget -qO- https://github.com/percona/mongodb_exporter/releases/download/v0.49.0/mongodb_exporter-0.49.0.linux-"${TARGETARCH}".tar.gz | tar xvzf - -C /opt/ --strip-components=1;

# Add the run.sh script
ADD ./run.sh /opt/

# Make the run.sh script executable
RUN chmod +x /opt/run.sh

# Set the user to 1001
USER 1001
