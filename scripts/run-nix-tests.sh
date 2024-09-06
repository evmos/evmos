#!/bin/sh
echo "check if gomod2nix.toml is updated with latest changes in go.mod"
gomod2nix generate

# Check if there is a git diff for gomod2nix.toml
if git diff --quiet gomod2nix.toml; then
	echo "All good! gomod2nix.toml file is updated."
else
	echo "Error: There are changes in the go.mod file. You need to regenerate the gomod2nix.toml file with the command 'gomod2nix generate'"
	exit 0 # Exit and don't run the tests. Exit with code 0 so the GH actions CI updates the gomod2nix file programmatically
fi

set -e
cd "$(dirname "$0")"

# explicitly set a short TMPDIR to prevent path too long issue on macosx
export TMPDIR=/tmp

echo "build test contracts"
cd ../tests/nix_tests/hardhat
npm install
npm run typechain
cd ..

# we want to pass the arguments as separate words instead of one string
# shellcheck disable=SC2086
pytest $ARGS -vv -s
