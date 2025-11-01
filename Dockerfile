FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine3.22 AS builder

ENV GOSUMDB=off GOPRIVATE=github.com/Netcracker

RUN apk add --no-cache git
RUN --mount=type=secret,id=GH_ACCESS_TOKEN \
    git config --global url."https://$(cat /run/secrets/GH_ACCESS_TOKEN)@github.com/".insteadOf "https://github.com/"

COPY . /workspace

WORKDIR /workspace
RUN go mod tidy

# Build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o ./build/_output/bin/mongodb-operator \
    -gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} ./main.go

# Copy files for Helm deployment
RUN mkdir -p deployments/charts/mongodb-operator && \
    cp -R ./charts/helm/mongodb-operator/* deployments/charts/mongodb-operator/ && \
    cp ./charts/deployment-configuration.json deployments/deployment-configuration.json

FROM alpine:3.22.1

ENV OPERATOR=/usr/local/bin/mongodb-operator \
    USER_UID=1001 \
    USER_NAME=mongodb-operator

RUN echo 'https://dl-cdn.alpinelinux.org/alpine/v3.22/main/' > /etc/apk/repositories \
    && apk add --no-cache openssl curl \
    && apk update \
    && apk upgrade
    
# RUN apk add --no-cache \
#   --repository=https://dl-cdn.alpinelinux.org/alpine/edge/main \
#   libcurl curl

# install operator binary
COPY --from=builder /workspace/build/_output/bin/mongodb-operator ${OPERATOR}
COPY build/bin /usr/local/bin

RUN chmod +x /usr/local/bin/entrypoint
RUN  chmod +x /usr/local/bin/user_setup && /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
