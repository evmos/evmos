#!/usr/bin/env bash
# ------------------
# Paths
#
COSMOS_URL=https://raw.githubusercontent.com/cosmos/cosmos-sdk/main
ETHERMINT_URL=https://github.com/evmos/ethermint
IBC_GO_URL=https://raw.githubusercontent.com/cosmos/ibc-go/main
# Formatting script
FORMAT=./format/format_cosmos_specs.py

# ------------------
# Setup
#
rm -rf modules && mkdir -p modules

for D in ../x/*; do
  if [ -d "${D}" ]; then
    rm -rf "modules/$(echo "$D" | awk -F/ '{print $NF}')"
    mkdir -p "modules/$(echo "$D" | awk -F/ '{print $NF}')" && cp -r "$D"/spec/* "$_"
  fi
done

sed 's/\.\/x/\/modules/g' ../x/README.md | sed 's/spec\/README.md//g' | sed 's/\.\.\/docs\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./modules/README.md

# ------------------
# Include the specs from Ethermint
#
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

# ------------------
# Include the specs from Cosmos SDK
#
# NOTE: Using curl to get Cosmos specs, because there is always only one file per folder.
#       This is much quicker.
mkdir ./modules/auth
curl -sSL "$COSMOS_URL"/x/auth/README.md > ./modules/auth/README.md
curl -sSL "$COSMOS_URL"/x/auth/vesting/README.md > ./modules/auth/vesting.md
curl -sSL "$COSMOS_URL"/x/auth/tx/README.md > ./modules/auth/tx.md
$FORMAT ./modules/auth/README.md --header --order 1
$FORMAT ./modules/auth/vesting.md --header --order 2
$FORMAT ./modules/auth/tx.md --header --order 3

mkdir ./modules/bank
curl -sSL "$COSMOS_URL"/x/bank/README.md > ./modules/bank/README.md
$FORMAT ./modules/bank/README.md --header

mkdir ./modules/crisis
curl -sSL "$COSMOS_URL"/x/crisis/README.md > ./modules/crisis/README.md
$FORMAT ./modules/crisis/README.md --header

mkdir ./modules/distribution
curl -sSL "$COSMOS_URL"/x/distribution/README.md > ./modules/distribution/README.md
$FORMAT ./modules/distribution/README.md --header

mkdir ./modules/evidence
curl -sSL "$COSMOS_URL"/x/evidence/README.md > ./modules/evidence/README.md
$FORMAT ./modules/evidence/README.md --header

mkdir ./modules/gov
curl -sSL "$COSMOS_URL"/x/gov/README.md > ./modules/gov/README.md
$FORMAT ./modules/gov/README.md --header

mkdir ./modules/slashing
curl -sSL "$COSMOS_URL"/x/slashing/README.md > ./modules/slashing/README.md
$FORMAT ./modules/slashing/README.md --header

mkdir ./modules/staking
curl -sSL "$COSMOS_URL"/x/staking/README.md > ./modules/staking/README.md
$FORMAT ./modules/staking/README.md --header

mkdir ./modules/upgrade
curl -sSL "$COSMOS_URL"/x/upgrade/README.md > ./modules/upgrade/README.md
$FORMAT ./modules/upgrade/README.md --header

# ------------------
# Include the specs from IBC go
#
# NOTE: no need to create the modules/ibc directory because it is already created in
#       the for loop at beginning of the script.
curl -sSL "$IBC_GO_URL"/modules/core/spec/01_concepts.md > ./modules/ibc/01_concepts.md
