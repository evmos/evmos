<!--
order: 7
-->

# Clients

A user can query the `x/claims` module using the CLI, gRPC or REST.

## CLI

Find below a list of `evoblockd` commands added with the `x/claims` module. You can obtain the full list by using the `evoblockd -h` command.

### Queries

The `query` commands allow users to query `claims` state.

**`total-unclaimed`**

Allows users to query total amount of unclaimed tokens from the airdrop.

```bash
evoblockd query claims total-unclaimed [flags]
```

**`records`**

Allows users to query all the claims records available.

```bash
evoblockd query claims records [flags]
```

**`record`**

Allows users to query a claims record for a given user.

```bash
evoblockd query claims record [address] [flags]
```

**`params`**

Allows users to query claims params.

```bash
evoblockd query claims params [flags]
```

## gRPC

### Queries

| Verb   | Method                                     | Description                                      |
|--------|--------------------------------------------|--------------------------------------------------|
| `gRPC` | `evoblock.claims.v1.Query/TotalUnclaimed`     | Gets the total unclaimed tokens from the airdrop |
| `gRPC` | `evoblock.claims.v1.Query/ClaimsRecords`      | Gets all registered claims records               |
| `gRPC` | `evoblock.claims.v1.Query/ClaimsRecord`       | Get the claims record for a given user            |
| `gRPC` | `evoblock.claims.v1.Query/Params`             | Gets claims params                               |
| `GET`  | `/evoblock/claims/v1/total_unclaimed`         | Gets the total unclaimed tokens from the airdrop |
| `GET`  | `/evoblock/claims/v1/claims_records`          | Gets all registered claims records               |
| `GET`  | `/evoblock/claims/v1/claims_records/{address}` | Gets a claims record for a given user            |
| `GET`  | `/evoblock/claims/v1/params`                  | Gets claims params                               |
