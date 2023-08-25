{ sources ? import ./sources.nix, system ? builtins.currentSystem, ... }:

import sources.nixpkgs {
  overlays = [
    (final: pkgs: rec {
      go_1_20 = pkgs.go_1_20.overrideAttrs (_: rec {
        version = "1.20.2";
        src = final.fetchurl {
          url = "https://go.dev/dl/go${version}.src.tar.gz";
          hash = "sha256-TQ4oUNGXtN2tO9sBljABedCVuzrv1N+8OzZwLDco+Ks=";
        };
      });
      go = go_1_20;
      go-ethereum = pkgs.callPackage ./go-ethereum.nix {
        inherit (pkgs.darwin) libobjc;
        inherit (pkgs.darwin.apple_sdk.frameworks) IOKit;
        buildGoModule = pkgs.buildGo118Module;
      };
    }) # update to a version that supports eip-1559
    # https://github.com/NixOS/nixpkgs/pull/179622
    (final: prev:
      (import "${sources.gomod2nix}/overlay.nix")
        (final // {
          inherit (final.darwin.apple_sdk_11_0) callPackage;
        })
        prev)
    (pkgs: _:
      import ./scripts.nix {
        inherit pkgs;
        config = {
          evmos-config = ../scripts/evmos-devnet.yaml;
          geth-genesis = ../scripts/geth-genesis.json;
          dotenv = builtins.path { name = "dotenv"; path = ../scripts/.env; };
        };
      })
    (_: pkgs: { test-env = pkgs.callPackage ./testenv.nix { }; })
  ];
  config = { };
  inherit system;
}
