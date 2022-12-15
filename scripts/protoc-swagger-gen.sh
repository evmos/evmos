#!/usr/bin/env bash

set -eo pipefail

# Settings
export SWAGGER_YAML="proto/buf.gen.swagger.yaml"
export SWAGGER_FOLDER="./tmp-swagger-gen"

# need to install statik on the docker image
go install github.com/rakyll/statik

# create temporary folder to store intermediate results from `buf generate`
mkdir -p "$SWAGGER_FOLDER"

# create swagger files on an individual basis  with `buf generate` (needed for `swagger-combine`)
proto_dirs=$(find ./proto ./third_party -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ -n "$query_file" ]]; then
    buf generate --template "$SWAGGER_YAML" "$query_file"
  fi
done

# Generate Swagger files for Ethermint proto files.
echo "Build ethermint files"
buf generate --template "$SWAGGER_YAML" "buf.build/evmos/ethermint"

# Generate Swagger files for Cosmos proto files.
echo "Build cosmos-proto files"
buf generate --template "$SWAGGER_YAML" "buf.build/cosmos/cosmos-proto"

echo "Build ibc files"
buf generate --template "$SWAGGER_YAML" "buf.build/cosmos/ibc"

# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf "$SWAGGER_FOLDER"

# generate binary for static server (use -f flag to replace current binary)
statik -f -src=./client/docs/swagger-ui -dest=./client/docs
