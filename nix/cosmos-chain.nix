{ lib
, buildGo119Module
, src
, version
, pname
, rev
, vendorSha256
}:
buildGo119Module rec {
  # Use this nix file to build any cosmos chain you need,
  # e.g. Stride, Osmosis, etc.
  inherit src version pname rev vendorSha256;
  tags = [ "netgo" ];
  ldflags = lib.concatStringsSep "\n" ([
    "-X github.com/cosmos/cosmos-sdk/version.Name=stride"
    "-X github.com/cosmos/cosmos-sdk/version.AppName=${pname}"
    "-X github.com/cosmos/cosmos-sdk/version.Version=${version}"
    "-X github.com/cosmos/cosmos-sdk/version.BuildTags=${lib.concatStringsSep "," tags}"
    "-X github.com/cosmos/cosmos-sdk/version.Commit=${rev}"
  ]);

  doCheck = false;
  subPackages = [ "cmd/${pname}" ];
}
