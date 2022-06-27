package types

import (
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	claimstypes "github.com/evmos/evmos/v6/x/claims/types"
)

// BankKeeper defines the banking keeper that must be fulfilled when
// creating a x/recovery keeper.
type BankKeeper interface {
	IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(coin sdk.Coin) (stop bool))
	BlockedAddr(addr sdk.AccAddress) bool
}

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
}

// TransferKeeper defines the expected IBC transfer keeper.
type TransferKeeper interface {
	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (transfertypes.DenomTrace, bool)
	SendTransfer(
		ctx sdk.Context,
		sourcePort, sourceChannel string,
		token sdk.Coin,
		sender sdk.AccAddress, receiver string,
		timeoutHeight clienttypes.Height, timeoutTimestamp uint64,
	) error
}

// ChannelKeeper defines the expected IBC channel keeper.
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
}

// ClaimsKeeper defines the expected claims keeper.
type ClaimsKeeper interface {
	GetParams(ctx sdk.Context) claimstypes.Params
}
