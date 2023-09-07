# Test Contracts

This directory contains the contracts that are used on the nix setup tests.
Its sole purpose is to use Hardhat to compile the contracts based on
the solidity compiler defined on the `hardhat.config.js` file.
Once compiled, the tests use the compiled data stored in the `artifacts`
directory to deploy and interact with the contracts.

To compile the contracts manually run:

```shell
npm install
npm run typechain
```

If you inspect the `package.json` file, you will notice that
the `typechain` command calls the `get-contracts` script.
This script copies all the Solidity smart contracts from the `precompiles`
directory of the evmos repository.
Thus, you don't need to add these contracts to the `contracts` directory,
these will be automatically included for you to use them on tests.
