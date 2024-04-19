{ system ? builtins.currentSystem, pkgs ? import ../../nix { inherit system; } }:
pkgs.mkShell {
  buildInputs = [
    pkgs.jq
    pkgs.go
    pkgs.gomod2nix
    (pkgs.callPackage ../../. { })
    pkgs.start-scripts
    pkgs.go-ethereum
    pkgs.nodejs
    pkgs.test-env
    pkgs.rocksdb
    pkgs.evmosd-rocksdb
    pkgs.chain-maind
    pkgs.strided
    pkgs.osmosisd
    pkgs.gaiad
    pkgs.hermes    
  ];
  shellHook = ''
    . ${../../scripts/.env}
  '';
}
