{ sources ? import ./sources.nix, system ? builtins.currentSystem,... }:
let
  # use a newer version of nixpkgs to get go_1_22
  # We're not updating this on the whole setup because breaks other stuff
  # but we can import the needed packages from the newer version
  # This is getting a specific commit on main branch to get
  # go v1.22 and rust v1.77. Using rust >= v1.78 breaks the current
  # hermes relayer compilation
  nixpkgsUrl = "https://github.com/NixOS/nixpkgs/archive/d1c6a5decfd9ad4c84354612d418b2856a57be1d.tar.gz";
  nixpkgs = import (fetchTarball nixpkgsUrl) {};
  # the go_1_22 nixpkgs is v1.22.1
  # but we need the v1.22.8.
  # This overrides the pkg to use
  # the v1.22.8 version
  go_1_22 = nixpkgs.pkgs.go_1_22.overrideAttrs {
    pname = "golang";
    version = "go1.22.8";
    src = nixpkgs.fetchFromGitHub {
      owner = "golang";
      repo = "go";
      rev = "aeccd613c896d39f582036aa52917c85ecf0b0c0";
      sha256 = "sha256-N3uG+FLMgThIAr1aDJSq+X+VKCz8dw6az35um3Mr3D0=";
    };
  };
  # get the rustPlatform used to build the hermes relayer
  # This rustPlatform uses rust v1.77
  rustPlatform = nixpkgs.pkgs.rustPlatform;
  gaiadBinUrl = if system == "x86_64-linux" then
      "https://github.com/cosmos/gaia/releases/download/v11.0.0/gaiad-v11.0.0-linux-amd64"
    else if system == "aarch64-darwin" then
      "https://github.com/cosmos/gaia/releases/download/v11.0.0/gaiad-v11.0.0-darwin-arm64"
    else
      throw "Unsupported architecture: ${system}";
  gaiadSha256 = if system == "x86_64-linux" then
      "sha256-JY3y7sWyL4uq3JiOGE+/0q5vn4iOn0RhoRDMNl/oYwA="
    else if system == "aarch64-darwin" then
      "sha256-U9D/5Ng1PlHQvlQ+33ZN4DPiTXA9TECCRKFB5jWydig="
    else
      throw "Unsupported architecture: ${system}";
in
import sources.nixpkgs {
  overlays = [
    (final: pkgs: rec {
      go = go_1_22;
      go-ethereum = pkgs.callPackage ./go-ethereum.nix {
        inherit (pkgs.darwin) libobjc;
        inherit (pkgs.darwin.apple_sdk.frameworks) IOKit;
        buildGoModule = pkgs.buildGo118Module;
      };
      rocksdb = pkgs.callPackage ./rocksdb.nix {};
      # evmos with rocksdb build
      evmosd-rocksdb = pkgs.callPackage ../default.nix { dbBackend = "rocksdb"; };
      # other chains to use in IBC tests
      chain-maind = pkgs.callPackage sources.chain-main { rocksdb = null; };
      strided = pkgs.callPackage ./cosmos-chain.nix { 
        src = sources.stride; 
        name = "stride";
        appName = "strided";
        version = "v16.0.0";
        rev = "e0c02910e036f4f2894a96c5222aebacc3ce0a4a";
        vendorSha256 = "sha256-vktJQOnnr/QcxiReMnCrlKEFqarMMFzfMjoB3LQ27vk=";
        patches = [ ../tests/nix_tests/configs/stride-admins.patch ]; # patch stride to allow tests addresses perform transactions that would need a gov proposal instead
      };
      # In case of osmosis & gaia, they provide the compiled binary. We'll use this
      # cause it is faster than building from source
      osmosisd = pkgs.callPackage ./bin.nix {
        appName = "osmosisd";
        version = "v20.2.1";
        binUrl = "https://github.com/osmosis-labs/osmosis/releases/download/v20.2.1/osmosisd-20.2.1-linux-amd64";
        sha256 = "sha256-TmCocIYcoXgZ+8tJ//mBtXMewRIdfLq0OYfF8E/wmfo=";
      };
      # Using gaia v11 (includes the PFM) cause after this version the '--min-self-delegation' flag is removed
      # from the 'gentx' cmd. 
      # This is needed cause pystarport has this hardcoded when spinning up the 
      # the environment
      gaiad = pkgs.callPackage ./bin.nix {
        appName = "gaiad";
        version = "v11.0.0";
        binUrl = gaiadBinUrl;
        sha256 = gaiadSha256;
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
      hermes = pkgs.callPackage ./hermes.nix { src = sources.hermes; platform = rustPlatform; };
    })
    (_: pkgs: { test-env = pkgs.callPackage ./testenv.nix { }; })
  ];
  config = { };
  inherit system;
}
