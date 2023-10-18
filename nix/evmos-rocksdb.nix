{ lib
, buildGoApplication
, buildPackages
, stdenv
, rocksdb
, static ? stdenv.hostPlatform.isStatic
, rev ? "dirty"
}:
let
  version = "latest-rocksdb";
  pname = "evmosd";
  tags = [ "ledger" "netgo" "rocksdb" "grocksdb_clean_link"  ];
  ldflags = lib.concatStringsSep "\n" ([
    "-X github.com/cosmos/cosmos-sdk/version.Name=evmos"
    "-X github.com/cosmos/cosmos-sdk/version.AppName=${pname}"
    "-X github.com/cosmos/cosmos-sdk/version.Version=${version}"
    "-X github.com/cosmos/cosmos-sdk/version.BuildTags=${lib.concatStringsSep "," tags}"
    "-X github.com/cosmos/cosmos-sdk/version.Commit=${rev}"
    "-X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb"
  ]);
  buildInputs = [ rocksdb ];
in
buildGoApplication rec {
  inherit pname version buildInputs tags ldflags;
  go = buildPackages.go_1_20;
  src = ../.;
  modules = ../gomod2nix.toml;
  doCheck = false;
  pwd = src; # needed to support replace
  subPackages = [ "cmd/evmosd" ];
  CGO_ENABLED = "1";

  postFixup = ''
    # Rename the binary from evmosd to evmosd-rocksdb
    mv $out/bin/evmosd $out/bin/evmosd-rocksdb
  '';

  meta = with lib; {
    description = "Evmos is a scalable and interoperable blockchain, built on Proof-of-Stake with fast-finality using the Cosmos SDK which runs on top of CometBFT Core consensus engine.";
    homepage = "https://github.com/evmos/evmos";
    license = licenses.asl20;
    mainProgram = "evmosd-rocksdb";
  };
}
