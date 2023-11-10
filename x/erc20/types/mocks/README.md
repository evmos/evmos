# Mocks

The mocks in this folder have been generated using the [mockery](https://vektra.github.io/mockery/latest/) tool.
To regenerate the mocks, run the following commands:

- `BankKeeper` (from used version of Cosmos SDK):

```bash
git clone https://github.com/evmos/cosmos-sdk.git
cd cosmos-sdk
git checkout v0.47.5 # or the version currently used

# Go into bank module and generate mock
cd x/bank
mockery --name Keeper
```

- `EVMKeeper` (reduced interface defined in ERC20 types):

```bash
cd x/erc20/types
mockery --name EVMKeeper
```
