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
      # other chains to use in IBC tests
      chain-maind = pkgs.callPackage sources.chain-main { rocksdb = null; };
      strided = pkgs.callPackage ./cosmos-chain.nix { 
        src = sources.stride; 
        name = "stride";
        appName = "strided";
        version = "v11.0.0";
        rev = "4b5d80ac5cafb418debc8a860959d4a6c6797cfb";
        vendorSha256 = "sha256-x3jAEsq/eWkPdyoDwFwARa7XeLxUj7t6hjScxeGoP/0=";
      };
      # In case of osmosis & gaia, they provide the compiled binary. We'll use this
      # cause it is faster than building from source
      osmosisd = pkgs.callPackage ./bin.nix {
        appName = "osmosisd";
        version = "v19.2.0";
        binUrl = "https://github.com/osmosis-labs/osmosis/releases/download/v19.2.0/osmosisd-19.2.0-linux-amd64";
        sha256 = "sha256-cj/xxTSes8A5w9xfVYlbveLhSZ/nwKlpYMxvre7IFMQ=";
      };
      # Using gaia v11 (includes the PFM) cause after this version the '--min-self-delegation' flag is removed
      # from the 'gentx' cmd. 
      # This is needed cause pystarport has this hardcoded when spinning up the 
      # the environment
      gaiad = pkgs.callPackage ./bin.nix {
        appName = "gaiad";
        version = "v11.0.0";
        binUrl = "https://github.com/cosmos/gaia/releases/download/v11.0.0/gaiad-v11.0.0-linux-amd64";
        sha256 = "sha256-JY3y7sWyL4uq3JiOGE+/0q5vn4iOn0RhoRDMNl/oYwA=";
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
    (import (fetchTarball "https://github.com/oxalica/rust-overlay/archive/master.tar.gz"))
    (_: pkgs: {
      hermes = pkgs.callPackage ./hermes.nix { src = sources.hermes; };
    })
    (_: pkgs: { test-env = pkgs.callPackage ./testenv.nix { }; })
  ];
  config = { };
  inherit system;
}
