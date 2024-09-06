#!/bin/bash
export GOPATH="$HOME"/go
export PATH="$PATH":"$GOPATH"/bin

# remove existing data
rm -rf "$HOME"/.evmosd

# build evmos binary
make install

cd tests/solidity || exit

if command -v yarn &>/dev/null; then
	yarn install
else
	curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
	echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
	sudo apt update && sudo apt install yarn
	yarn install
fi

yarn test --network evmos "$@"
