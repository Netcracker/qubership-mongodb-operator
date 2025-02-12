FROM golang:1.22.5-alpine3.19

ENV OPERATOR=/usr/local/bin/qubership-mongodb-operator \
    USER_UID=1001 \
    USER_NAME=qubership-mongodb-operator

ENV WORKDIR=/opt/operator/
ENV GOSUMDB=off GOPRIVATE=github.com/Netcracker/*

# Set up Git access (optional if you're using private repos)
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


ENV WORKDIR=/opt/operator/

# Set working directory
WORKDIR ${WORKDIR}

# Copy necessary source files into the container
COPY charts /opt/operator/charts
COPY main.go /opt/operator/main.go


# Build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o ./bin/qubership-mongodb-operator \
-gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} ./main.go


# Copy files for Helm deployment
RUN mkdir -p deployments/charts/mongodb-operator && \
    cp -R ./charts/helm/mongodb-operator/* deployments/charts/mongodb-operator/ && \
    cp ./charts/deployment-configuration.json deployments/deployment-configuration.json

# Install the Go binary and set up entrypoint
COPY bin/qubership-mongodb-operator /usr/local/bin/qubership-mongodb-operator
COPY build/bin /usr/local/bin

# Set permissions for entrypoint and user setup
RUN chmod +x /usr/local/bin/entrypoint && \
    chmod +x /usr/local/bin/user_setup && /usr/local/bin/user_setup

# Set entrypoint for the container
ENTRYPOINT ["/usr/local/bin/entrypoint"]

# Set the user for the container (ensure you have a user setup earlier)
USER ${USER_UID}
