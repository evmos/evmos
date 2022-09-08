<!--
order: 3
-->

# Chain ID

Learn about the Point Chain chain-id format {synopsis}

## Official Chain IDs

::: tip
**NOTE**: The latest Chain ID (i.e highest Version Number) is the latest version of the software and mainnet.
:::

:::: tabs
::: tab Mainnet

| Name                                            | Chain ID                                      | Identifier | EIP155 Number                         | Version Number                              |
| ----------------------------------------------- | --------------------------------------------- | ---------- | ------------------------------------- | ------------------------------------------- |
| Point Chain 1                                         | `point_{{ $themeConfig.project.chain_id }}-1` | `point`    | `{{ $themeConfig.project.chain_id }}` | `1`                                         |
:::
::: tab Testnets

| Name                              | Chain ID                                              | Identifier | EIP155 Number                                 | Version Number                                      |
| --------------------------------- | ----------------------------------------------------- | ---------- | --------------------------------------------- | --------------------------------------------------- |
| Point Chain Public Testnet Uranus            | `point_{{ $themeConfig.project.testnet_chain_id }}-1` | `point`    | `{{ $themeConfig.project.testnet_chain_id }}` | `1`                                                 |
:::
::::

::: tip
You can also lookup the [EIP155](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md) `Chain ID` by referring to [chainlist.org](https://chainlist.org/).
:::

![chainlist.org website](./../../img/chainlist.png)

## The Chain Identifier

Every chain must have a unique identifier or `chain-id`. Tendermint requires each application to
define its own `chain-id` in the [genesis.json fields](https://docs.tendermint.com/master/spec/core/genesis.html#genesis-fields). However, in order to comply with both EIP155 and Cosmos standard for chain upgrades, Point Chain-compatible chains must implement a special structure for their chain identifiers.

## Structure

The Point Chain Chain ID contains 3 main components

- **Identifier**: Unstructured string that defines the name of the application.
- **EIP155 Number**: Immutable [EIP155](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md) `CHAIN_ID` that defines the replay attack protection number.
- **Version Number**: Is the version number (always positive) that the chain is currently running.
This number **MUST** be incremented every time the chain is upgraded or forked in order to avoid network or consensus errors.

### Format

The format for specifying and Point Chain compatible chain-id in genesis is the following:

```bash
{identifier}_{EIP155}-{version}
```

The following table provides an example where the second row corresponds to an upgrade from the first one:

| ChainID        | Identifier | EIP155 Number | Version Number |
| -------------- | ---------- | ------------- | -------------- |
| `point_10687-1` | point      | 10687          | 1              |
| `point_10687-2` | point      | 10687          | 2              |
| `...`          | ...        | ...           | ...            |
| `point_10687-N` | point      | 10687          | N              |
