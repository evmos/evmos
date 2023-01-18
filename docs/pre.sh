#!/usr/bin/env bash
# ------------------
# Paths
#
COSMOS_URL=https://raw.githubusercontent.com/cosmos/cosmos-sdk/main
ETHERMINT_URL=https://github.com/evmos/ethermint
IBC_GO_URL=https://github.com/cosmos/ibc-go.git
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
# For this purpose we are using the sparse checkout and only pull the following folders:
#   - x/evm/spec
#   - x/feemarket/spec
#
# Additionally, we are applying formatting to the feemarket overview file to
# match the rest of the Evmos docs.
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
# Include the specs from Cosmos SDK and apply formatting to all downloaded files
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
# For this purpose we are using the sparse checkout and only pull the following folders:
#   - docs/apps/transfer
#   - docs/apps/interchain-accounts
#   - docs/assets/ica
#
# Additionally, we are deleting the "legacy" subfolder from the transfer module and apply
# formatting to the overview files (adjusting header text and adding file metadata).
#
# NOTE: no need to create the modules/ibc directory because it is already created in
#       the for loop at beginning of the script.
mkdir ibc_specs
cd ibc_specs || exit
git init
git remote add origin "$IBC_GO_URL"
git config core.sparseCheckout true
printf "docs/apps/transfer\ndocs/apps/interchain-accounts\ndocs/assets/ica\n" > .git/info/sparse-checkout
git pull origin main
cd ..

mv ibc_specs/docs/apps/transfer/ ./modules/transfer
mv ibc_specs/docs/apps/interchain-accounts/ ./modules/interchain-accounts
mkdir assets && mkdir assets/ica
mv ibc_specs/docs/assets/ica/ica-v6.png ./assets/ica/ica-v6.png
rm -rf ibc_specs
rm -rf ./modules/interchain-accounts/legacy
sed 's/\# Overview/\# transfer/' ./modules/transfer/overview.md > ./modules/transfer/overview_tmp.md
mv ./modules/transfer/overview_tmp.md ./modules/transfer/overview.md
$FORMAT ./modules/transfer/overview.md --header --order 0 --title "Transfer Overview" --parent "transfer"
sed 's/\# Overview/\# interchain-accounts/' ./modules/interchain-accounts/overview.md > ./modules/interchain-accounts/overview_tmp.md
mv ./modules/interchain-accounts/overview_tmp.md ./modules/interchain-accounts/overview.md
$FORMAT ./modules/interchain-accounts/overview.md --header --order 0 --title "ICA Overview" --parent "interchain-accounts"
