#!/usr/bin/env bash

rm -rf modules && mkdir -p modules

for D in ../x/*; do
  if [ -d "${D}" ]; then
    rm -rf "modules/$(echo $D | awk -F/ '{print $NF}')"
    mkdir -p "modules/$(echo $D | awk -F/ '{print $NF}')" && cp -r $D/spec/* "$_"
  fi
done

cat ../x/README.md | sed 's/\.\/x/\/modules/g' | sed 's/spec\/README.md//g' | sed 's/\.\.\/docs\/building-modules\/README\.md/\/building-modules\/intro\.html/g' > ./modules/README.md

# Include the evm spec from Ethermint
git clone https://github.com/tharsis/ethermint.git
mv ethermint/x/evm/spec/ ./modules/evm && rm -rf ethermint