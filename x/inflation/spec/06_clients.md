<!--
order: 8
-->

# Clients

A user can query the `x/incentives` module using the CLI, JSON-RPC, gRPC or
REST.

## CLI

Find below a list of `berachaind` commands added with the `x/inflation` module. You
can obtain the full list by using the `berachaind -h` command.

### Queries


The `query` commands allow users to query `inflation` state.

**`period`**

Allows users to query the current inflation period.

```go
berachaind query inflation period [flags]
```

**`epoch-mint-provisions`**

Allows users to query the current inflation epoch provisions value.

```go
berachaind query inflation epoch-mint-provisions [flags]
```

**`params`**

Allows users to query the current inflation parameters.

```go
berachaind query inflation params [flags]
```

### Proposals

The `tx gov submit-proposal` commands allow users to query create a proposal
using the governance module CLI:

**`register-incentive`**

Allows users to submit a `RegisterIncentiveProposal`.

```bash
berachaind tx gov submit-proposal register-incentive [contract-address] [allocation] [epochs] [flags]
```

**`cancel-incentive`**

Allows users to submit a `CanelIncentiveProposal`.

```bash
berachaind tx gov submit-proposal cancel-incentive [contract-address] [flags]
```

**`param-change`**

Allows users to submit a `ParameterChangeProposal`.

```bash
berachaind tx gov submit-proposal param-change [proposal-file] [flags]
```

## gRPC

### Queries

| Verb   | Method                                        | Description                                   |
| ------ | --------------------------------------------- | --------------------------------------------- |
| `gRPC` | `evmos.inflation.v1.Query/Period`             | Gets current inflation period                 |
| `gRPC` | `evmos.inflation.v1.Query/EpochMintProvision` | Gets current inflation epoch provisions value |
| `gRPC` | `evmos.inflation.v1.Query/Params`             | Gets current inflation parameters             |
| `GET`  | `/evmos/inflation/v1/period`                  | Gets current inflation period                 |
| `GET`  | `/evmos/inflation/v1/epoch_mint_provision`    | Gets current inflation epoch provisions value |
| `GET`  | `/evmos/inflation/v1/params`                  | Gets current inflation parameters             |
