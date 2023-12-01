{ pkgs ? import ../../../nix { } }:
let evmosd = (pkgs.callPackage ../../../. { });
in
evmosd.overrideAttrs (oldAttrs: {
  # Patch the evmos binary to:
  # - allow to register WEVMOS token pair
  # - use the CrossChainSwap contract address in the testing setup
  # - update the corresponding IBC channel to match the tests setup
  patches = oldAttrs.patches or [ ] ++ [
    ./allow-wevmos-register.patch
    ./xcs-osmosis-contract.patch
    ./osmosis-channel.patch
  ];
})

