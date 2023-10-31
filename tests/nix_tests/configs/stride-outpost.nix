{ pkgs ? import ../../../nix { } }:
let evmosd = (pkgs.callPackage ../../../. { });
in
evmosd.overrideAttrs (oldAttrs: {
  # Patch the evmos binary to:
  # - allow to register WEVMOS token pair
  # - use channel-0 for the stride outpost
  patches = oldAttrs.patches or [ ] ++ [
    ./allow-wevmos-register.patch
    ./stride-outpost-channel.patch
  ];
})
