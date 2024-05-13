{ lib
, buildGoApplication
, buildPackages
, fetchFromGitHub
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
  # use a newer version of nixpkgs to get go_1_22
  # We're not updating this on the whole setup because breaks other stuff
  # but we can import the needed packages from the newer version
  nixpkgsUrl = "https://github.com/NixOS/nixpkgs/archive/master.tar.gz";
  nixpkgs = import (fetchTarball nixpkgsUrl) {};
  # the go_1_22 nixpkgs is v1.22.1
  # but we need the v1.22.2. 
  # This overrides the pkg to use
  # the v1.22.2 version  
  go_1_22 = nixpkgs.pkgs.go_1_22.overrideAttrs {
    pname = "golang";
    version = "go1.22.2";
    src = fetchFromGitHub {
      owner = "golang";
      repo = "go";
      rev = "dddf0ae40fa0c1223aba191d73a44425a08e1035";
      sha256 = "sha256-gWJ4txAt2TkobDo1EGotWDOSP2pGqLCNqpn+Smgr21w=";
    };
  };  
in
buildGoApplication rec {
  inherit pname version buildInputs tags ldflags;
  go = go_1_22;
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
