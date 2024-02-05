FROM golang:1.22rc2-bullseye as build-env

# Install minimum necessary dependencies
ENV PACKAGES curl make git libc-dev bash gcc
# Ignoring the lint about pinning version because
# the build breaks whenever a new version comes out
# hadolint ignore=DL3008
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y $PACKAGES --no-install-recommends

# Set working directory for the build
WORKDIR /go/src/github.com/evmos/evmos

# Add source files
COPY . .

# build Ethermint
RUN make build-linux

# Final image
FROM golang:1.22rc2-bullseye as final

WORKDIR /

RUN apt-get update

# Copy over binaries from the build-env
COPY --from=build-env /go/src/github.com/evmos/evmos/build/evmosd /
COPY --from=build-env /go/src/github.com/evmos/evmos/scripts/start-docker.sh /

EXPOSE 26656 26657 1317 9090 8545 8546

# Run evmosd by default, omit entrypoint to ease using container with evmosd
ENTRYPOINT ["/bin/bash", "-c"]