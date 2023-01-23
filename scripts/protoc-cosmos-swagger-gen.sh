#!/usr/bin/env bash

set -eo pipefail

mkdir -p ./tmp-swagger-gen
cd ./third_party/proto
proto_dirs=$(find ./cosmos -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ -n "$query_file" ]]; then
    echo "Generating $query_file"
    # buf generate --template ../buf.gen.cosmos-swagger.yaml "$query_file"
    buf generate --template ./buf.gen.swagger.yaml "$query_file"
  fi
done
cd ../..
