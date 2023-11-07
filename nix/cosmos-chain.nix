{ lib
, buildGoModule
, buildPackages
, src
, version
, name
, appName
, rev
, vendorSha256
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

  CGO_ENABLED = "1";
  GOWORK = "off";

  doCheck = false;
  subPackages = [ "cmd/${appName}" ];
}
