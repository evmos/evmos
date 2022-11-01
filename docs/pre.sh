#!/usr/bin/env bash

rm -rf modules && mkdir -p modules

for D in ../x/*; do
  if [ -d "${D}" ]; then
    rm -rf "modules/$(echo $D | awk -F/ '{print $NF}')"
    mkdir -p "modules/$(echo $D | awk -F/ '{print $NF}')" && cp -r $D/spec/* "$_"
  fi
done

cat ../x/README.md | sed 's/\.\/x/\/modules/g' | sed 's/spec\/README.md//g' | sed 's/\.\.\/docs\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./modules/README.md

# Include the specs from Ethermint
git clone https://github.com/evoblockchain/ethermint.git
mv ethermint/x/evm/spec/ ./modules/evm 
mv ethermint/x/feemarket/spec/ ./modules/feemarket 
rm -rf ethermint

# Include the specs from Cosmos SDK
git clone https://github.com/cosmos/cosmos-sdk.git
mv cosmos-sdk/x/auth/spec/ ./modules/auth
mv cosmos-sdk/x/bank/spec/ ./modules/bank
mv cosmos-sdk/x/crisis/spec/ ./modules/crisis
mv cosmos-sdk/x/distribution/spec/ ./modules/distribution
mv cosmos-sdk/x/evidence/spec/ ./modules/evidence
mv cosmos-sdk/x/gov/spec/ ./modules/gov
mv cosmos-sdk/x/slashing/spec/ ./modules/slashing
mv cosmos-sdk/x/staking/spec/ ./modules/staking
mv cosmos-sdk/x/upgrade/spec/ ./modules/upgrade
rm -rf cosmos-sdk

# Include the specs from IBC go
git clone https://github.com/cosmos/ibc-go.git
mv ibc-go/modules/apps/transfer/spec/ ./modules/transfer
mv ibc-go/modules/core/spec/ ./modules/ibc-core
rm -rf ibc-go