{ lib
, buildGoApplication
, buildPackages
, rocksdb
, rev ? "dirty"
}:
let
  version = "latest";
  pname = "evmosd";
  tags = [ "ledger" "netgo" "rocksdb" "grocksdb_no_link"  ];
  ldflags = lib.concatStringsSep "\n" ([
    "-X github.com/cosmos/cosmos-sdk/version.Name=evmos"
    "-X github.com/cosmos/cosmos-sdk/version.AppName=${pname}"
    "-X github.com/cosmos/cosmos-sdk/version.Version=${version}-rocksdb"
    "-X github.com/cosmos/cosmos-sdk/version.BuildTags=${lib.concatStringsSep "," tags}"
    "-X github.com/cosmos/cosmos-sdk/version.Commit=${rev}"
    "-X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb"
  ]);
in
buildGoApplication rec {
  inherit pname version tags ldflags;
  go = buildPackages.go_1_20;
  src = ../.;
  modules = ./gomod2nix.toml;
  doCheck = false;
  pwd = src; # needed to support replace
  subPackages = [ "cmd/evmosd" ];
  CGO_ENABLED = "1";
  CGO_LDFLAGS =
    if static then "-lrocksdb -pthread -lstdc++ -ldl -lzstd -lsnappy -llz4 -lbz2 -lz"
    else if stdenv.hostPlatform.isWindows then "-lrocksdb-shared"
    else "-lrocksdb -pthread -lstdc++ -ldl";

  postFixup = lib.optionalString stdenv.isDarwin ''
    ${stdenv.cc.targetPrefix}install_name_tool -change "@rpath/librocksdb.8.dylib" "${rocksdb}/lib/librocksdb.dylib" $out/bin/cronosd
  '';

  meta = with lib; {
    description = "Evmos is a scalable and interoperable blockchain, built on Proof-of-Stake with fast-finality using the Cosmos SDK which runs on top of CometBFT Core consensus engine.";
    homepage = "https://github.com/evmos/evmos";
    license = licenses.asl20;
    mainProgram = "evmosd";
  };
}
