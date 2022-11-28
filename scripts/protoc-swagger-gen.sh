#!/usr/bin/env bash

set -eo pipefail

# Settings
export SWAGGERYAML="proto/buf.gen.swagger.yaml"
export SWAGGERFOLDER="./tmp-swagger-gen"

#buf ls-files "buf.build/evmos/ethermint"

## need to install statik on the docker image
#go install github.com/rakyll/statik
#
## create temporary folder to store intermediate results from `buf build` + `buf generate`
#mkdir -p "$SWAGGERFOLDER"

# create swagger files on an individual basis  w/ `buf build` and `buf generate` (needed for `swagger-combine`)
proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ -n "$query_file" ]]; then
    buf generate --template "$SWAGGERYAML" "$query_file"
  fi
done

## NOTE: This works because there is only one package contained in cosmos/cosmos-proto.
## Generate Swagger files for third party and Ethermint proto files.
#echo "Build cosmos-proto files"
#buf generate --template "$SWAGGERYAML" "buf.build/cosmos/cosmos-proto"

# FIXME: These commands don't work yet. When only using the top level evmos/ethermint registry
# an error occurs saying that there are inconsistent package names.
# TODO: Find way to specify exact package to generate in remote input
echo "Build ethermint files"
#buf generate --template "$SWAGGERYAML" "buf.build/evmos/ethermint"
buf generate --template "$SWAGGERYAML" "https://github.com/evmos/ethermint.git#tag=v0.19.3,subdir=proto/ethermint/crypto"
#buf generate --template "$SWAGGERYAML" "buf.build/evmos/ethermint/crypto/v1/ethsecp256k1"
#buf generate --template "$SWAGGERYAML" "buf.build/evmos/ethermint/ethermint/crypto"
#buf generate --template "$SWAGGERFOLDER" "buf.build/evmos/ethermint#subdir=ethermint/crypto"
#
#echo "Build cosmos-sdk files"
#buf generate --template "$SWAGGERYAML" "buf.build/cosmos/cosmos-sdk"

## combine swagger files
## uses nodejs package `swagger-combine`.
## all the individual swagger files need to be configured in `config.json` for merging
#swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true
#
## clean swagger files
#rm -rf "$SWAGGERFOLDER"
#
## generate binary for static server (use -f flag to replace current binary)
#statik -f -src=./client/docs/swagger-ui -dest=./client/docs
