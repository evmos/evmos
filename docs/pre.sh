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
ETHERMINT_URL=https://github.com/evmos/ethermint

mkdir ethermint_specs
cd ethermint_specs || exit
git init
git remote add origin "$ETHERMINT_URL"
git config core.sparseCheckout true
printf "x/evm/spec\nx/feemarket/spec\n" > .git/info/sparse-checkout
git pull origin main
ls
cd ..

mv ethermint_specs/x/evm/spec/ ./modules/evm
mv ethermint_specs/x/feemarket/spec/ ./modules/feemarket
rm -rf ethermint_specs

# Include the specs from Cosmos SDK
COSMOS_URL=https://github.com/cosmos/cosmos-sdk

mkdir cosmos_sdk_specs
cd cosmos_sdk_specs || exit
git init
git remote add origin "$COSMOS_URL"
git config core.sparseCheckout true
printf "x/auth/spec\nx/bank/spec\nx/crisis/spec\nx/distribution/spec\nx/evidence/spec\nx/gov/spec\nx//spec\nx/slashing/spec\nx/staking/spec\nx/upgrade/spec\n" > .git/info/sparse-checkout
git pull origin main
ls
cd ..

mv cosmos_sdk_specs/x/auth/spec/ ./modules/auth
mv cosmos_sdk_specs/x/bank/spec/ ./modules/bank
mv cosmos_sdk_specs/x/crisis/spec/ ./modules/crisis
mv cosmos_sdk_specs/x/distribution/spec/ ./modules/distribution
mv cosmos_sdk_specs/x/evidence/spec/ ./modules/evidence
mv cosmos_sdk_specs/x/gov/spec/ ./modules/gov
mv cosmos_sdk_specs/x/slashing/spec/ ./modules/slashing
mv cosmos_sdk_specs/x/staking/spec/ ./modules/staking
mv cosmos_sdk_specs/x/upgrade/spec/ ./modules/upgrade
rm -rf cosmos_sdk_specs

# Include the specs from IBC go
IBC_GO_URL=https://github.com/cosmos/ibc-go

mkdir ibc-go
cd ibc-go || exit
git init
git remote add origin "$IBC_GO_URL"
git config core.sparseCheckout true
printf "modules/core/spec\n" > .git/info/sparse-checkout
git pull origin main
ls
cd ..

mv ibc-go/modules/core/spec/ ./modules/ibc-core
rm -rf ibc-go
