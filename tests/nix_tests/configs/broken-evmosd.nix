{ pkgs ? import ../../../nix { } }:
let eidond = (pkgs.callPackage ../../../. { });
in
eidond.overrideAttrs (oldAttrs: {
  patches = oldAttrs.patches or [ ] ++ [
    ./broken-eidond.patch
  ];
})
