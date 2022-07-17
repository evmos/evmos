#!/usr/bin/env bash

set -eo pipefail

protoc_gen_go_pulsar() {
  if ! grep "github.com/gogo/protobuf => github.com/regen-network/protobuf" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the evmos folder."
    return 1
  fi

  go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@v1.0.0-alpha7
}

protoc_gen_doc() {
  go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest 2>/dev/null
}

protoc_gen_go_pulsar
protoc_gen_doc

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
# TODO: migrate to `buf build`
for dir in $proto_dirs; do
  buf alpha protoc \
  -I "proto" \
  -I "third_party/proto" \
  --go-pulsar_out=. --go-pulsar_opt=paths=source_relative $(find "${dir}" -maxdepth 1 -name '*.proto')
done

# command to generate docs using protoc-gen-doc
# TODO: migrate to `buf build`
buf alpha protoc \
-I "proto" \
-I "third_party/proto" \
--doc_out=./docs/protocol \
--doc_opt=./docs/protodoc-markdown.tmpl,proto-docs.md \
$(find "$(pwd)/proto" -maxdepth 5 -name '*.proto')

# move proto files to the right places
cp -r github.com/evmos/evmos/v*/x/* x/
rm -rf github.com
