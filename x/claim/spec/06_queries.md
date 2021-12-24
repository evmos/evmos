<!--
order: 6
-->

# Queries

## GRPC queries

Claim module provides below GRPC queries to query claim status

```protobuf
service Query {
  rpc ModuleAccountBalance(QueryModuleAccountBalanceRequest) returns (QueryModuleAccountBalanceResponse) {}
  rpc Params(QueryParamsRequest) returns (QueryParamsResponse) {}
  rpc ClaimRecord(QueryClaimRecordRequest) returns (QueryClaimRecordResponse) {}
  rpc ClaimableForAction(QueryClaimableForActionRequest) returns (QueryClaimableForActionResponse) {}
  rpc TotalClaimable(QueryTotalClaimableRequest) returns (QueryTotalClaimableResponse) {}
}
```

## CLI commands

For the following commands, you can change `$(osmosisd keys show -a {your key name})` with the address directly.

Query the claim record for a given address

```sh
osmosisd query claim claim-record $(osmosisd keys show -a {your key name})
```

Query the claimable amount that would be earned if a specific action is completed right now.

```sh

osmosisd query claim claimable-for-action $(osmosisd keys show -a {your key name}) ActionAddLiquidity
```

Query the total claimable amount that would be earned if all remaining actions were completed right now.
Note that even if the decay process hasn't begun yet, this is not always *exactly* the same as `InitialClaimableAmount`, due to rounding errors.

```sh
osmosisd query claim total-claimable $(osmosisd keys show -a {your key name}) ActionAddLiquidity
```
