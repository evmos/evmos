FROM golang:1.21.4-alpine3.18 AS build-env

ARG DB_BACKEND=goleveldb
ARG ROCKSDB_VERSION="8.5.3"

WORKDIR /go/src/github.com/evmos/evmos

COPY go.mod go.sum ./

RUN set -eux; apk add --no-cache ca-certificates=20230506-r0 build-base=0.5-r3 git=2.40.1-r0 linux-headers=6.3-r0 bash=5.2.15-r5

RUN --mount=type=bind,target=. --mount=type=secret,id=GITHUB_TOKEN \
    git config --global url."https://$(cat /run/secrets/GITHUB_TOKEN)@github.com/".insteadOf "https://github.com/"; \
    go mod download

COPY . .

RUN mkdir -p /target/usr/lib /target/usr/local/lib /target/usr/include

RUN if [ "$DB_BACKEND" = "pebbledb" ]; then \
    make build-pebbledb; \
elif [ "$DB_BACKEND" = "rocksdb" ]; then \
   make build-rocksdb; \
   cp -r /usr/lib/* /target/usr/lib/ && \
   cp -r /usr/local/lib/* /target/usr/local/lib/ && \
   cp -r /usr/include/* /target/usr/include/; \
else \
    # Build default binary (LevelDB)
    make build; \
fi

RUN go install github.com/MinseokOh/toml-cli@latest

FROM alpine:3.18

WORKDIR /root

COPY --from=build-env /go/src/github.com/evmos/evmos/build/evmosd /usr/bin/evmosd
COPY --from=build-env /go/bin/toml-cli /usr/bin/toml-cli

# required for rocksdb build
COPY --from=build-env /target/usr/lib /usr/lib
COPY --from=build-env /target/usr/local/lib /usr/local/lib
COPY --from=build-env /target/usr/include /usr/include

RUN apk add --no-cache ca-certificates=20230506-r0 jq=1.6-r4 curl=8.5.0-r0 bash=5.2.15-r5 vim=9.0.2073-r0 lz4=1.9.4-r4 rclone=1.62.2-r6 \
    && addgroup -g 1000 evmos \
    && adduser -S -h /home/evmos -D evmos -u 1000 -G evmos

USER 1000
WORKDIR /home/evmos

EXPOSE 26656 26657 1317 9090 8545 8546

CMD ["evmosd"]
