<!--
order: 2
-->

# State

## State Objects

The `x/vesting` module does not keep objects in its own store. Instead, it uses the SDK `auth` module to store account objects in state using the [Account Interface](https://docs.cosmos.network/main/modules/auth#account-interface). Accounts are exposed externally as an interface and stored internally as a clawback vesting account.

## ClawbackVestingAccount

An instance that implements the [Vesting Account](https://docs.cosmos.network/main/modules/auth/vesting#vesting-account-types) interface. It provides an account that can hold contributions subject to lockup, or vesting which is subject to clawback of unvested tokens, or a combination (tokens vest, but are still locked).

```go
type ClawbackVestingAccount struct {
	// base_vesting_account implements the VestingAccount interface. It contains
	// all the necessary fields needed for any vesting account implementation
	*types.BaseVestingAccount `protobuf:"bytes,1,opt,name=base_vesting_account,json=baseVestingAccount,proto3,embedded=base_vesting_account" json:"base_vesting_account,omitempty"`
	// funder_address specifies the account which can perform clawback
	FunderAddress string `protobuf:"bytes,2,opt,name=funder_address,json=funderAddress,proto3" json:"funder_address,omitempty"`
	// start_time defines the time at which the vesting period begins
	StartTime time.Time `protobuf:"bytes,3,opt,name=start_time,json=startTime,proto3,stdtime" json:"start_time"`
	// lockup_periods defines the unlocking schedule relative to the start_time
	LockupPeriods []types.Period `protobuf:"bytes,4,rep,name=lockup_periods,json=lockupPeriods,proto3" json:"lockup_periods"`
	// vesting_periods defines the vesting schedule relative to the start_time
	VestingPeriods []types.Period `protobuf:"bytes,5,rep,name=vesting_periods,json=vestingPeriods,proto3" json:"vesting_periods"`
}
```

### BaseVestingAccount

Implements the `VestingAccount` interface. It contains all the necessary fields needed for any vesting account implementation.

### FunderAddress

Specifies the account which provides the original tokens and can perform clawback.

### StartTime

Defines the time at which the vesting and lockup schedules begin.

### LockupPeriods

Defines the unlocking schedule relative to the start time.

### VestingPeriods

Defines the vesting schedule relative to the start time.

## Genesis State

The `x/vesting` module allows the definition of `ClawbackVestingAccounts` at genesis. In this case, the account balance must be logged in the SDK `bank` module balances or automatically adjusted through the `add-genesis-account` CLI command.
