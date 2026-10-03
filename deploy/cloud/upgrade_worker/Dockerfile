# The upgrade worker runs with the host Docker socket and a bind-mounted
# checkout. Keep the image tool-only so workflow scripts stay reviewable in
# source and the cloud host never installs packages at container startup.
FROM alpine:3.22@sha256:3e9b4b680bfc9fb5269227cffbd6d42be39fbf7c0b908123913864aa4447e764

RUN apk add --no-cache \
    bash \
    ca-certificates \
    curl \
    docker-cli \
    docker-cli-compose \
    git \
    jq \
    util-linux

