# Compiling Smart Contracts

This tool enables to generated the JSON files with compiled smart contracts,
that can be found in this repository.
Previously, the compilation and building of JSON files
that were not contained in the `contracts` directory
had to be done manually with Remix, which was very tedious and error-prone.

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

**Note**: The tool will compile all smart contracts found
(except for the ignored paths defined in the script)
but only overwrite the compiled JSON data for contracts
that already have a corresponding compiled JSON file in the same directory.

If you want to add a new smart contract and have a JSON file generated for it,
run:

```bash
make contracts-add CONTRACT=path/to/contract.sol
```
