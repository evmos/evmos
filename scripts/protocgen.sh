#!/usr/bin/env bash

# How to run manually:
# TODO: Change comments to Evmos
# docker build --pull --rm -f "contrib/devtools/Dockerfile" -t cosmossdk-proto:latest "contrib/devtools"
# docker run --rm -v $(pwd):/workspace --workdir /workspace cosmossdk-proto sh ./scripts/protocgen.sh

set -e

echo "Generating gogo proto code"

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    # TODO: Adjust this command for Evmos
    # this regex checks if a proto file has its go_package set to cosmossdk.io/api/...
    # gogo proto files SHOULD ONLY be generated if this is false
    # we don't want gogo proto to run for proto files which are natively built for google.golang.org/protobuf
    if grep -q "option go_package.*evmos" "$file"; then
      buf generate --template proto/buf.gen.gogo.yaml $file
    fi
  done
done

# move proto files to the right places
cp -r github.com/evmos/evmos/v*/x/* x/
rm -rf github.com
