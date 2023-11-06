FROM golang:1.21.3-alpine3.18 AS build-env

ARG DB_BACKEND=goleveldb

# This is only needed when building
# the binary with rocksdb
ARG ROCKSDB_VERSION=v8.5.3

WORKDIR /go/src/github.com/evmos/evmos

COPY go.mod go.sum ./

RUN set -eux; apk add --no-cache ca-certificates=20230506-r0 build-base=0.5-r3 git=2.40.1-r0 linux-headers=6.3-r0

RUN --mount=type=bind,target=. --mount=type=secret,id=GITHUB_TOKEN \
    git config --global url."https://$(cat /run/secrets/GITHUB_TOKEN)@github.com/".insteadOf "https://github.com/"; \
    go mod download

COPY . .

RUN mkdir -p /usr/local/rocksdb/lib
RUN if [ "$DB_BACKEND" = "rocksdb" ]; then \
    echo "Building with RocksDB support"; \
    # 1. Install dependencies
    echo "@testing http://nl.alpinelinux.org/alpine/edge/testing" >>/etc/apk/repositories; \
    apk add --update --no-cache cmake bash perl g++; \
    apk add --update --no-cache zlib zlib-dev bzip2 bzip2-dev snappy snappy-dev lz4 lz4-dev zstd@testing zstd-dev@testing libtbb-dev@testing libtbb@testing; \
    # 2. Install latest gflags
    cd /tmp && \
    git clone https://github.com/gflags/gflags.git && \
    cd gflags && \
    mkdir build && \
    cd build && \
    cmake -DBUILD_SHARED_LIBS=1 -DGFLAGS_INSTALL_SHARED_LIBS=1 .. && \
    make install; \
    # 3. Install Rocksdb
    cd /tmp && \
    git clone -b ${ROCKSDB_VERSION} --single-branch https://github.com/facebook/rocksdb.git && \
    cd rocksdb && \
    PORTABLE=1 WITH_JNI=0 WITH_BENCHMARK_TOOLS=0 WITH_TESTS=1 WITH_TOOLS=0 WITH_CORE_TOOLS=1 WITH_BZ2=1 WITH_LZ4=1 WITH_SNAPPY=1 WITH_ZLIB=1 WITH_ZSTD=1 WITH_GFLAGS=0 USE_RTTI=1 \
    make shared_lib && \
    mkdir /usr/local/rocksdb/include && \
    cp librocksdb.so* /usr/local/rocksdb/lib && \
    cp /usr/local/rocksdb/lib/librocksdb.so* /usr/lib/ && \
    cp -r include /usr/local/rocksdb/ && \
    cp -r include/* /usr/include/; \
    # 4. Set corresponding build flags & build evmosd binary
    cd /go/src/github.com/evmos/evmos && \
    CGO_ENABLED=1 CGO_CFLAGS="-I/usr/include" CGO_LDFLAGS="-L/usr/local/rocksdb -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -ldl" COSMOS_BUILD_OPTIONS=${DB_BACKEND} make build; \
elif [ "$DB_BACKEND" = "pebbledb" ]; then \
    echo "Building with PebbleDB support"; \
    # Replace the cometbft-db dependency to support PebbleDB build
    go mod edit -replace github.com/cometbft/cometbft-db=github.com/notional-labs/cometbft-db@pebble && \
    go mod tidy; \
    # Build with PebbleDB support
    COSMOS_BUILD_OPTIONS=${DB_BACKEND} make build; \
else \
    # Build default binary (LevelDB)
    make build; \
fi

RUN go install github.com/MinseokOh/toml-cli@latest

FROM alpine:3.18

WORKDIR /root

COPY --from=build-env /go/src/github.com/evmos/evmos/build/evmosd /usr/bin/evmosd
COPY --from=build-env /go/bin/toml-cli /usr/bin/toml-cli

# These are required for rocksdb build
COPY --from=build-env /usr/lib /usr/lib
COPY --from=build-env /usr/include /usr/include
COPY --from=build-env /usr/local/lib /usr/local/lib
COPY --from=build-env /usr/local/rocksdb /usr/local/rocksdb

RUN apk add --no-cache ca-certificates=20230506-r0 jq=1.6-r3 curl=8.4.0-r0 bash=5.2.15-r5 vim=9.0.1568-r0 lz4=1.9.4-r4 rclone=1.62.2-r5 \
    && addgroup -g 1000 evmos \
    && adduser -S -h /home/evmos -D evmos -u 1000 -G evmos

USER 1000
WORKDIR /home/evmos

EXPOSE 26656 26657 1317 9090 8545 8546

CMD ["evmosd"]
