<!--
order: 4
-->

# Keepers

## Keeper functions

Claim keeper module provides utility functions to manage epochs.

```go
  GetModuleAccountAddress(ctx sdk.Context) sdk.AccAddress
  GetModuleAccountBalance(ctx sdk.Context) sdk.Coin
  EndAirdrop(ctx sdk.Context) error
  GetClaimRecord(ctx sdk.Context, addr sdk.AccAddress) (types.ClaimRecord, error)
  GetClaimRecords(ctx sdk.Context) []types.ClaimRecord
  SetClaimRecord(ctx sdk.Context, claimRecord types.ClaimRecord) error
  SetClaimRecords(ctx sdk.Context, claimRecords []types.ClaimRecord) error
  GetClaimableAmountForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Coins, error)
  GetUserTotalClaimable(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error)
  ClaimCoinsForAction(ctx sdk.Context, addr sdk.AccAddress, action types.Action) (sdk.Coins, error)
  CreateModuleAccount(ctx sdk.Context, amount sdk.Coin)
  clearInitialClaimables(ctx sdk.Context)
  fundRemainingsToCommunity(ctx sdk.Context) error
```
