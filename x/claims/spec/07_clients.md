<!--
order: 7
-->

# Clients

A user can query the `x/claims` module using the CLI, gRPC or REST.

## CLI

Find below a list of `evmosd` commands added with the `x/claims` module. You can obtain the full list by using the `evmosd -h` command.

### Queries

The `query` commands allow users to query `claims` state.

```

**`claims-records`**

Allows users to query all the claim records available.

```bash
evmosd query claims claims-records [flags]
```

**`claims-record`**

Allows users to query a claims record for a given user.

```go
evmosd query claims claims-record [address] [flags]
```

**`params`**

Allows users to query claims params.

```bash
evmosd query claims params [flags]
```

## gRPC

### Queries

| Verb   | Method                                                     | Description                                   |
| ------ | ---------------------------------------------------------- | --------------------------------------------- |
| `gRPC` | `evmos.claims.v1.Query/Incentives`                     | Gets all registered claims                |
| `gRPC` | `evmos.claims.v1.Query/Incentive`                      | Gets incentive for a given contract           |
| `gRPC` | `evmos.claims.v1.Query/GasMeters`                      | Gets gas meters for a given incentive         |
| `gRPC` | `evmos.claims.v1.Query/GasMeter`                       | Gets gas meter for a given incentive and user |
| `gRPC` | `evmos.claims.v1.Query/AllocationMeters`               | Gets all allocation meters                    |
| `gRPC` | `evmos.claims.v1.Query/AllocationMeter`                | Gets allocation meter for a denom             |
| `gRPC` | `evmos.claims.v1.Query/Params`                         | Gets claims params                        |
| `GET`  | `/evmos/claims/v1/claims`                          | Gets all registered claims                |
| `GET`  | `/evmos/claims/v1/claims/{contract}`               | Gets incentive for a given contract           |
| `GET`  | `/evmos/claims/v1/gas_meters`                          | Gets gas meters for a given incentive         |
| `GET`  | `/evmos/claims/v1/gas_meters/{contract}/{participant}` | Gets gas meter for a given incentive and user |
| `GET`  | `/evmos/claims/v1/allocation_meters`                   | Gets all allocation meters                    |
| `GET`  | `/evmos/claims/v1/allocation_meters/{denom}`           | Gets allocation meter for a denom             |
| `GET`  | `/evmos/claims/v1/params`                              | Gets claims params                        |
