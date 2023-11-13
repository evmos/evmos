{ pkgs ? import ../../../nix { } }:
let evmosd = (pkgs.callPackage ../../../. { });
in
evmosd.overrideAttrs (oldAttrs: {
  # Patch the evmos binary to:
  # - allow to register WEVMOS token pair
  # - use the CrossChainSwap contract address in the testing setup
  patches = oldAttrs.patches or [ ] ++ [
    ./allow-wevmos-register.patch
    ./xcs-osmosis-contract.patch
  ];
})

