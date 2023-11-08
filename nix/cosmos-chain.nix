{ lib
, buildGoModule
, buildPackages
, src
, version
, name
, appName
, rev
, vendorSha256
, wasmEnabled ? false
}:
buildGoModule rec {
  go = buildPackages.go_1_20;
  # Use this nix file to build any cosmos chain you need,
  # e.g. Stride, Osmosis, etc.
  inherit src version name appName rev vendorSha256;
  tags = [ "netgo" ];

  ldflags = lib.concatStringsSep "\n" ([
    "-X github.com/cosmos/cosmos-sdk/version.Name=${name}"
    "-X github.com/cosmos/cosmos-sdk/version.AppName=${appName}"
    "-X github.com/cosmos/cosmos-sdk/version.Version=${version}"
    "-X github.com/cosmos/cosmos-sdk/version.BuildTags=${lib.concatStringsSep "," tags}"
    "-X github.com/cosmos/cosmos-sdk/version.Commit=${rev}"
  ]);

    # Install libwasm if required for the chain
  postUnpack = if wasmEnabled then ''
    cd $src
    go version
    ls
    uname=$(uname -m) && \
    go mod download && \
    WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc."$uname".a \
    -O $out/libwasmvm_muslc.a && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum $out/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep libwasmvm_muslc."$uname" | cut -d ' ' -f 1)
 '' else '''';

  CGO_ENABLED = "1";
  GOWORK = "off";

  doCheck = false;
  subPackages = [ "cmd/${appName}" ];
}
