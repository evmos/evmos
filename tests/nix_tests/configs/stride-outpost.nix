{ pkgs ? import ../../../nix { } }:
let evmosd = (pkgs.callPackage ../../../. { });
in
evmosd.overrideAttrs (oldAttrs: {
  # Patch the evmos binary to:
  # - use channel-0 for the stride outpost
  patches = oldAttrs.patches or [ ] ++ [
    ./stride-outpost-channel.patch
  ];
})
