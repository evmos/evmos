package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
)

func BlockedMessages() []string {
	return []string{
		sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
		sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
		sdk.MsgTypeURL(&sdkvesting.MsgCreatePermanentLockedAccount{}),
		sdk.MsgTypeURL(&sdkvesting.MsgCreatePeriodicVestingAccount{}),
	}
}
