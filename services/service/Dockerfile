FROM --platform=$BUILDPLATFORM golang:1.26.1-alpine3.23 AS builder

ENV GOSUMDB=off GOPRIVATE=github.com/Netcracker

RUN apk add --no-cache git
RUN --mount=type=secret,id=GH_ACCESS_TOKEN \
    git config --global url."https://$(cat /run/secrets/GH_ACCESS_TOKEN)@github.com/".insteadOf "https://github.com/"

COPY . /workspace

WORKDIR /workspace
RUN go mod tidy

# Build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o ./build/_output/bin/mongodb-services \
    -gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} ./main.go

# Copy files for Helm deployment
RUN mkdir -p deployments/charts/mongodb-services && \
    gzip -f -c ./charts/helm/mongodb-services/monitoring/mongo7-grafana-dashboard.json > ./charts/helm/mongodb-services/monitoring/mongo7-grafana-dashboard.json.gz && \
    gzip -f -c ./charts/helm/mongodb-services/monitoring/mongo8-grafana-dashboard.json > ./charts/helm/mongodb-services/monitoring/mongo8-grafana-dashboard.json.gz && \
    cp -R ./charts/helm/mongodb-services/* deployments/charts/mongodb-services/ && \
    cp ./charts/deployment-configuration.json deployments/deployment-configuration.json 

FROM alpine:3.23

ENV OPERATOR=/usr/local/bin/mongodb-services \
    USER_UID=1001 \
    USER_NAME=mongodb-services
    
RUN apk add --no-cache openssl curl

RUN echo "https://dl-cdn.alpinelinux.org/alpine/edge/main" > /etc/apk/repositories  \
    && echo "https://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories \
    && apk add --update --upgrade --no-cache \
        libcurl curl
        
# install operator binary
COPY --from=builder /workspace/build/_output/bin/mongodb-services ${OPERATOR}
COPY build/bin /usr/local/bin

RUN chmod +x /usr/local/bin/entrypoint
RUN  chmod +x /usr/local/bin/user_setup && /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
