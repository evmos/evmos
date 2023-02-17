<!--
order: 4
-->

# Transactions

This section defines the concrete `sdk.Msg`  types that result in the state transitions defined on the previous section.

## `CreateClawbackVestingAccount`

```go
type MsgCreateClawbackVestingAccount struct {
	// from_address specifies the account to provide the funds and sign the
	// clawback request
	FromAddress string `protobuf:"bytes,1,opt,name=from_address,json=fromAddress,proto3" json:"from_address,omitempty"`
	// to_address specifies the account to receive the funds
	ToAddress string `protobuf:"bytes,2,opt,name=to_address,json=toAddress,proto3" json:"to_address,omitempty"`
	// start_time defines the time at which the vesting period begins
	StartTime time.Time `protobuf:"bytes,3,opt,name=start_time,json=startTime,proto3,stdtime" json:"start_time"`
	// lockup_periods defines the unlocking schedule relative to the start_time
	LockupPeriods []types.Period `protobuf:"bytes,4,rep,name=lockup_periods,json=lockupPeriods,proto3" json:"lockup_periods"`
	// vesting_periods defines thevesting schedule relative to the start_time
	VestingPeriods []types.Period `protobuf:"bytes,5,rep,name=vesting_periods,json=vestingPeriods,proto3" json:"vesting_periods"`
	// merge specifies a the creation mechanism for existing
	// ClawbackVestingAccounts. If true, merge this new grant into an existing
	// ClawbackVestingAccount, or create it if it does not exist. If false,
	// creates a new account. New grants to an existing account must be from the
	// same from_address.
	Merge bool `protobuf:"varint,6,opt,name=merge,proto3" json:"merge,omitempty"`
}
```

The msg content stateless validation fails if:

- `FromAddress` or `ToAddress` are invalid
- `LockupPeriods` and `VestingPeriods`
    - include period with a non-positive length
    - describe the same total amount

## `Clawback`

```go
type MsgClawback struct {
	// funder_address is the address which funded the account
	FunderAddress string `protobuf:"bytes,1,opt,name=funder_address,json=funderAddress,proto3" json:"funder_address,omitempty"`
	// account_address is the address of the ClawbackVestingAccount to claw back from.
	AccountAddress string `protobuf:"bytes,2,opt,name=account_address,json=accountAddress,proto3" json:"account_address,omitempty"`
	// dest_address specifies where the clawed-back tokens should be transferred
	// to. If empty, the tokens will be transferred back to the original funder of
	// the account.
	DestAddress string `protobuf:"bytes,3,opt,name=dest_address,json=destAddress,proto3" json:"dest_address,omitempty"`
}
```

The msg content stateless validation fails if:

- `FunderAddress` or `AccountAddress` are invalid
- `DestAddress` is not empty and invalid

## `UpdateVestingFunder`

```go
type MsgUpdateVestingFunder struct {
	// funder_address is the current funder address of the ClawbackVestingAccount
	FunderAddress string `protobuf:"bytes,1,opt,name=funder_address,json=funderAddress,proto3" json:"funder_address,omitempty"`
	// new_funder_address is the new address to replace the existing funder_address
	NewFunderAddress string `protobuf:"bytes,2,opt,name=new_funder_address,json=newFunderAddress,proto3" json:"new_funder_address,omitempty"`
	// vesting_address is the address of the ClawbackVestingAccount being updated
	VestingAddress string `protobuf:"bytes,3,opt,name=vesting_address,json=vestingAddress,proto3" json:"vesting_address,omitempty"`
}
```

The msg content stateless validation fails if:

- `FunderAddress`, `NewFunderAddress` or `VestingAddress` are invalid

## `ConvertVestingAccount`

```go
type MsgConvertVestingAccount struct {
	// vesting_address is the address of the ClawbackVestingAccount being updated
	VestingAddress string `protobuf:"bytes,2,opt,name=vesting_address,json=vestingAddress,proto3" json:"vesting_address,omitempty"`
}
```

The msg content stateless validation fails if:

- `VestingAddress` is invalid
