FROM golang:1.21.0-alpine3.18 AS build-env

WORKDIR /go/src/github.com/evmos/evmos

COPY go.mod go.sum ./

RUN set -eux; apk add --no-cache ca-certificates=20230506-r0 build-base=0.5-r3 git=2.40.1-r0 linux-headers=6.3-r0

RUN --mount=type=bind,target=. --mount=type=secret,id=GITHUB_TOKEN \
    git config --global url."https://$(cat /run/secrets/GITHUB_TOKEN)@github.com/".insteadOf "https://github.com/"; \
    go mod download

COPY . .

#RUN make build
RUN go mod edit -replace github.com/tendermint/tm-db=github.com/notional-labs/tm-db@v0.6.8-pebble \
    && go mod tidy \
    && go mod edit -replace github.com/cometbft/cometbft-db=github.com/notional-labs/cometbft-db@pebble \
    && go mod tidy \
    && go install -ldflags "-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=pebbledb \
       -X github.com/cosmos/cosmos-sdk/version.Version=$(git describe --tags)-pebbledb \
       -X github.com/cosmos/cosmos-sdk/version.Commit=$(git log -1 --format='%H')" -tags pebbledb ./...

RUN go install github.com/MinseokOh/toml-cli@latest

FROM alpine:3.18

WORKDIR /root

COPY --from=build-env /go/bin/evmosd /usr/bin/evmosd
COPY --from=build-env /go/bin/toml-cli /usr/bin/toml-cli

RUN apk add --no-cache ca-certificates=20230506-r0 jq=1.6-r3 curl=8.4.0-r0 bash=5.2.15-r5 vim=9.0.1568-r0 lz4=1.9.4-r4 rclone=1.62.2-r5 \
    && addgroup -g 1000 evmos \
    && adduser -S -h /home/evmos -D evmos -u 1000 -G evmos

USER 1000
WORKDIR /home/evmos

EXPOSE 26656 26657 1317 9090 8545 8546

CMD ["evmosd"]
