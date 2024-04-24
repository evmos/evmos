# Compiling Smart Contracts

This tool compiles all smart contracts found in this repository using a Hardhat setup.
The contracts are collected and then copied into the `contracts` directory for compilation.
After compilation, the resulting JSON data is copied back to the source locations.

**Note**: The tool will compile all smart contracts found
(except for the ignored paths defined in the script)
but only overwrite the compiled JSON data for contracts
that already have a corresponding compiled JSON file in the same directory.
If you want to add a new JSON file to the repository, use the `add` command
described below.

## Usage

To compile the smart contracts, run the following command:

```bash
make contracts-compile
```

This will compile the smart contracts and generate the JSON files.

To clean up the generated artifacts, installed dependencies and cached files,
run:

```bash
make contracts-clean
```

If you want to add a new smart contract and have a JSON file generated for it,
run:

```bash
make contracts-add CONTRACT=path/to/contract.sol
```
