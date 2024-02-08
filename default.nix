{ lib
, buildGoApplication
, buildPackages
, stdenv
, rev ? "dirty"
, rocksdb
, static ? stdenv.hostPlatform.isStatic
, dbBackend ? "goleveldb"
}:
let
  version = if dbBackend == "rocksdb" then "latest-rocksdb" else "latest";
  pname = "evmosd";
  tags = [ "ledger" "netgo" ] ++ lib.optionals (dbBackend == "rocksdb") [ "rocksdb" "grocksdb_clean_link" ];
  ldflags = lib.concatStringsSep "\n" ([
    "-X github.com/cosmos/cosmos-sdk/version.Name=evmos"
    "-X github.com/cosmos/cosmos-sdk/version.AppName=${pname}"
    "-X github.com/cosmos/cosmos-sdk/version.Version=${version}"
    "-X github.com/cosmos/cosmos-sdk/version.BuildTags=${lib.concatStringsSep "," tags}"
    "-X github.com/cosmos/cosmos-sdk/version.Commit=${rev}"
    "-X github.com/cosmos/cosmos-sdk/types.DBBackend=${dbBackend}"
  ]);
  buildInputs = lib.optionals (dbBackend == "rocksdb") [ rocksdb ];
  # use a newer version of nixpkgs to get go_1_21
  # We're not updating this on the whole setup because breaks other stuff
  # but we can import the needed packages from the newer version
  nixpkgsUrl = "https://github.com/NixOS/nixpkgs/archive/23.11.tar.gz";
  nixpkgs = import (fetchTarball nixpkgsUrl) {};
  go_1_21 = nixpkgs.pkgs.go_1_21;    
in
buildGoApplication rec {
  inherit pname version buildInputs tags ldflags;
  go = go_1_21;
  src = ./.;
  modules = ./gomod2nix.toml;
  doCheck = false;
  pwd = src; # needed to support replace
  subPackages = [ "cmd/evmosd" ];
  CGO_ENABLED = "1";

  postFixup = if dbBackend == "rocksdb" then
    ''
      # Rename the binary from evmosd to evmosd-rocksdb
      mv $out/bin/evmosd $out/bin/evmosd-rocksdb
    '' else '''';

  meta = with lib; {
    description = "Evmos is a scalable and interoperable blockchain, built on Proof-of-Stake with fast-finality using the Cosmos SDK which runs on top of CometBFT Core consensus engine.";
    homepage = "https://github.com/evmos/evmos";
    license = licenses.asl20;
    mainProgram = "evmosd";
  };
}
