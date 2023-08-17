let
  pkgs = import ../../../nix { };
  fetchEvmos = rev: builtins.fetchTarball "https://github.com/evmos/evmos/archive/${rev}.tar.gz";
  released = pkgs.buildGo118Module rec {
    name = "evmosd";
    src = fetchEvmos "92827302f11a33d01fb630d0d302075ddab361ae";
    subPackages = [ "cmd/evmosd" ];
    vendorSha256 = "sha256-cQAol54b6hNzsA4Q3MP9mTqFWM1MvR5uMPrYpaoj3SY=";
    doCheck = false;
  };
  current = pkgs.callPackage ../../../. { };
in
pkgs.linkFarm "upgrade-test-package" [
  { name = "genesis"; path = released; }
  { name = "integration-test-upgrade"; path = current; }
]
