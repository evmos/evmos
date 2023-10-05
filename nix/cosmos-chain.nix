{ lib
, buildGo119Module
, src
, version
, name
, appName
, rev
, vendorSha256
}:
buildGo119Module rec {
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

  doCheck = false;
  subPackages = [ "cmd/${appName}" ];
}
