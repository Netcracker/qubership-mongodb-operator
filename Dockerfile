FROM alpine:3.19.1

ENV WORKDIR=/opt/operator/
ENV GOSUMDB=off GOPRIVATE=github.com/Netcracker

# Set up Git access
RUN --mount=type=secret,id=GH_ACCESS_TOKEN git config --global url."https://$(cat /run/secrets/GH_ACCESS_TOKEN)@github.com/".insteadOf "https://github.com/"

# Install dependencies for Go build and script execution
RUN echo 'https://dl-cdn.alpinelinux.org/alpine/v3.19/main/' > /etc/apk/repositories \
    && apk add --no-cache \
    bash \
    git \
    curl \
    openssl \
    go \
    zip

# Set working directory in the container
WORKDIR ${WORKDIR}

# Copy the build script and the necessary files into the container
COPY build.sh /opt/operator/build.sh
COPY charts /opt/operator/charts
COPY main.go /opt/operator/main.go

# Make the build.sh script executable
RUN chmod +x /opt/operator/build.sh

# Run the build script
RUN /opt/operator/build.sh

# Copy the binary and other required files after the build
COPY bin/mongo-operator ${OPERATOR}
COPY build/bin /usr/local/bin

# Set permissions for entrypoint and user setup scripts
RUN chmod +x /usr/local/bin/entrypoint
RUN chmod +x /usr/local/bin/user_setup && /usr/local/bin/user_setup

# Set the entry point
ENTRYPOINT ["/usr/local/bin/entrypoint"]

# Set the user for the container
USER ${USER_UID}

