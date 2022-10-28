
## Build Image
FROM golang:1.18-bullseye as build

ARG PRE_UPGRADE_VERSION
ENV PRE_UPGRADE_VERSION $PRE_UPGRADE_VERSION

RUN apt-get update -y
RUN apt-get install git -y

# TODO: change origin
RUN git clone --depth 1 --branch $PRE_UPGRADE_VERSION https://github.com/evmos/evmos.git

WORKDIR /go/evmos/

RUN GO111MODULE=on go build -o ./build/chain_init ./tests/e2e/chain_init


## Deploy image
FROM ubuntu

COPY --from=build /go/evmos/build/chain_init /bin/chain_init

ENTRYPOINT [ "chain_init" ]
