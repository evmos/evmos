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
cd ..

mv ethermint_specs/x/evm/spec/ ./modules/evm
mv ethermint_specs/x/feemarket/spec/ ./modules/feemarket
rm -rf ethermint_specs
$FORMAT ./modules/feemarket/README.md --header --order 0 --title "Feemarket Overview" --parent "feemarket"

# ------------------
# Include the specs from Cosmos SDK
#
# NOTE: Using curl to get Cosmos specs, because there is always only one file per folder.
#       This is much quicker.
mkdir ./modules/auth
curl -sSL "$COSMOS_URL"/x/auth/README.md > ./modules/auth/README.md
curl -sSL "$COSMOS_URL"/x/auth/vesting/README.md > ./modules/auth/vesting.md
curl -sSL "$COSMOS_URL"/x/auth/tx/README.md > ./modules/auth/tx.md
$FORMAT ./modules/auth/README.md --header --order 1 --title "Auth Overview" --parent "auth"
$FORMAT ./modules/auth/vesting.md --header --order 2 --title "auth/vesting"
$FORMAT ./modules/auth/tx.md --header --order 3 --title "auth/tx"

mkdir ./modules/bank
curl -sSL "$COSMOS_URL"/x/bank/README.md > ./modules/bank/README.md
$FORMAT ./modules/bank/README.md --header --title "Bank Overview" --parent "bank"

mkdir ./modules/crisis
curl -sSL "$COSMOS_URL"/x/crisis/README.md > ./modules/crisis/README.md
$FORMAT ./modules/crisis/README.md --header --title "Crisis Overview" --parent "crisis"

mkdir ./modules/distribution
curl -sSL "$COSMOS_URL"/x/distribution/README.md > ./modules/distribution/README.md
$FORMAT ./modules/distribution/README.md --header --title "Distribution Overview" --parent "distribution"

mkdir ./modules/evidence
curl -sSL "$COSMOS_URL"/x/evidence/README.md > ./modules/evidence/README.md
$FORMAT ./modules/evidence/README.md --header --title "Evidence Overview" --parent "evidence"

mkdir ./modules/gov
curl -sSL "$COSMOS_URL"/x/gov/README.md > ./modules/gov/README.md
$FORMAT ./modules/gov/README.md --header --title "Gov Overview" --parent "gov"

mkdir ./modules/slashing
curl -sSL "$COSMOS_URL"/x/slashing/README.md > ./modules/slashing/README.md
$FORMAT ./modules/slashing/README.md --header --title "Slashing Overview" --parent "slashing"

mkdir ./modules/staking
curl -sSL "$COSMOS_URL"/x/staking/README.md > ./modules/staking/README.md
$FORMAT ./modules/staking/README.md --header --title "Staking Overview" --parent "staking"

mkdir ./modules/upgrade
curl -sSL "$COSMOS_URL"/x/upgrade/README.md > ./modules/upgrade/README.md
$FORMAT ./modules/upgrade/README.md --header --title "Upgrade Overview" --parent "upgrade"

# ------------------
# Include the specs from IBC go
#
# NOTE: no need to create the modules/ibc directory because it is already created in
#       the for loop at beginning of the script.
curl -sSL "$IBC_GO_URL"/docs/ibc/overview.md > ./modules/ibc/README.md
sed 's/\# Overview/\# ibc-go/' ./modules/ibc/README.md > ./modules/ibc/README_tmp.md
mv ./modules/ibc/README_tmp.md ./modules/ibc/README.md
$FORMAT ./modules/ibc/README.md --header --order 0 --title "IBC-Go Overview" --parent "ibc-go"

curl -sSL "$IBC_GO_URL"/docs/apps/transfer/overview.md > ./modules/ibc/transfer.md
sed 's/\# Overview/\# ibc-go\/transfer/' ./modules/ibc/transfer.md > ./modules/ibc/transfer_tmp.md
mv ./modules/ibc/transfer_tmp.md ./modules/ibc/transfer.md
$FORMAT ./modules/ibc/transfer.md --header --order 1 --title "ibc-go/transfer"

curl -sSL "$IBC_GO_URL"/docs/apps/interchain-accounts/overview.md > ./modules/ibc/interchain-accounts.md
sed 's/\# Overview/\# ibc-go\/interchain-accounts/' ./modules/ibc/interchain-accounts.md > ./modules/ibc/interchain-accounts_tmp.md
mv ./modules/ibc/interchain-accounts_tmp.md ./modules/ibc/interchain-accounts.md
$FORMAT ./modules/ibc/interchain-accounts.md --header --order 2 --title "ibc-go/interchain-accounts"
