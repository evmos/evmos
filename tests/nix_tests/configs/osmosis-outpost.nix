{ pkgs ? import ../../../nix { } }:
let evmosd = (pkgs.callPackage ../../../. { });
in
evmosd.overrideAttrs (oldAttrs: {
  # Patch the evmos binary to:
  # - use the CrossChainSwap contract address in the testing setup
  # - update the corresponding IBC channel to match the tests setup
  patches = oldAttrs.patches or [ ] ++ [
    ./osmosis-outpost.patch
  ];
})

