{ system ? builtins.currentSystem, pkgs ? import ../../nix { inherit system; } }:
let
  goEnv = pkgs.mkGoEnv { pwd = ../../.; };
in
pkgs.mkShell {
  buildInputs = [
    pkgs.jq
    pkgs.go
    pkgs.gomod2nix
    goEnv
    pkgs.start-scripts
    pkgs.go-ethereum
    pkgs.nodejs
    pkgs.test-env
    pkgs.gomod2nix
  ];
  shellHook = ''
    . ${../../scripts/.env}
  '';
}
