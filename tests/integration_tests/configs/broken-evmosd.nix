{ pkgs ? import ../../../nix { } }:
let evmosd = pkgs.mkGoEnv { pwd = ../../../.; };
in
evmosd.overrideAttrs (oldAttrs: {
  patches = oldAttrs.patches or [ ] ++ [
    ./broken-evmosd.patch
  ];
})
