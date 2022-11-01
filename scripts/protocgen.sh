#!/usr/bin/env bash

set -eo pipefail

protoc_gen_gocosmos() {
  if ! grep "github.com/gogo/protobuf => github.com/regen-network/protobuf" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the sommelier folder."
    return 1
  fi

  go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos 2>/dev/null
}

protoc_gen_doc() {
  go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc 2>/dev/null
}

protoc_gen_gocosmos
protoc_gen_doc

# need to install statik on the docker image
go install github.com/rakyll/statik

# create temporary folder to store intermediate results from `buf build` + `buf generate`
mkdir -p ./tmp-swagger-gen

# build .proto files and generate code for the proto/ directory
buf build proto
buf generate proto --template buf.gen.proto.yaml

# create additional swagger files on an individual basis  w/ `buf build` and `buf generate` (needed for `swagger-combine`)
proto_dirs=$(find ./proto ./third_party/proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do

  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    buf build --path "$query_file"
    buf generate --path "$query_file" --template buf.gen.swagger.yaml
  fi
done

# move resulting files to the right places
cp -r github.com/evoblockchain/evoblock/v*/x/* x/
rm -rf github.com

# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen

# generate binary for static server (use -f flag to replace current binary)
statik -f -src=./client/docs/swagger-ui -dest=./client/docs