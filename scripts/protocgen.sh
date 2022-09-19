#!/usr/bin/env bash
set -x
set -e
set -o pipefail

# get protoc executions
go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos 2>/dev/null

# get evmos from github
go get github.com/evmos/evmos/v8@v8.0.0 2>/dev/null
evmos_dir=$(go list -f '{{ .Dir }}' -m github.com/evmos/evmos/v8)

# Get the path of the cosmos-sdk repo from go/pkg/mod
proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

for dir in $proto_dirs; do
  # generate protobuf bind
  buf protoc \
  -I "proto" \
  -I "$evmos_dir/third_party/proto" \
  --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

  # generate grpc gateway
  buf protoc \
  -I "proto" \
  -I "$evmos_dir/third_party/proto" \
  --grpc-gateway_out=logtostderr=true:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')
done

# move proto files to the right places
cp -r github.com/ArableProtocol/acrechain/* ./
rm -rf github.com
