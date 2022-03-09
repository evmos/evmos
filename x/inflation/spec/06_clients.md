<!--
order: 8
-->

# Clients

A user can query the `x/incentives` module using the CLI, JSON-RPC, gRPC or
REST.

## CLI

Find below a list of `evmosd` commands added with the `x/inflation` module. You
can obtain the full list by using the `evmosd -h` command.

### Queries


The `query` commands allow users to query `inflation` state.

**`period`**

Allows users to query the current inflation period.

```go
evmosd query inflation period [flags]
```

**`epoch-mint-provisions`**

Allows users to query the current inflation epoch provisions value.

```go
evmosd query inflation epoch-mint-provisions [flags]
```

**`skipped-epochs`**

Allows users to query the current number of skipped epochs.

```go
evmosd query inflation skipped-epochs [flags]
```



**`params`**

Allows users to query the current inflation parameters.

```go
evmosd query inflation params [flags]
```

### Proposals

The `tx gov submit-proposal` commands allow users to query create a proposal
using the governance module CLI:

**`param-change`**

Allows users to submit a `ParameterChangeProposal`.

```bash
evmosd tx gov submit-proposal param-change [proposal-file] [flags]
```

## gRPC

### Queries

| Verb   | Method                                        | Description                                   |
| ------ | --------------------------------------------- | --------------------------------------------- |
| `gRPC` | `evmos.inflation.v1.Query/Period`             | Gets current inflation period                 |
| `gRPC` | `evmos.inflation.v1.Query/EpochMintProvision` | Gets current inflation epoch provisions value |
| `gRPC` | `evmos.inflation.v1.Query/Params`             | Gets current inflation parameters             |
| `gRPC` | `evmos.inflation.v1.Query/SkippedEpochs`      | Gets current number of skipped epochs         |
| `GET`  | `/evmos/inflation/v1/period`                  | Gets current inflation period                 |
| `GET`  | `/evmos/inflation/v1/epoch_mint_provision`    | Gets current inflation epoch provisions value |
| `GET`  | `/evmos/inflation/v1/skipped_epochs`          | Gets current number of skipped epochs         |
| `GET`  | `/evmos/inflation/v1/params`                  | Gets current inflation parameters             |
