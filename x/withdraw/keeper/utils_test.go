package keeper_test

import (
	"github.com/stretchr/testify/mock"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"

	"github.com/tharsis/evmos/v2/x/withdraw/types"
)

/*
  Test objects
*/

var _ types.TransferKeeper = &TransferKeeper{}

// TransferKeeper is a mocked object that implements an interface
// that describes an object that the code I am testing relies on.
type TransferKeeper struct {
	mock.Mock
	bankkeeper.Keeper
}

func (m *TransferKeeper) GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (transfertypes.DenomTrace, bool) {
	args := m.Called(denomTraceHash)
	return args.Get(0).(transfertypes.DenomTrace), args.Bool(1)
}

func (m *TransferKeeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	args := m.Called(sourcePort, sourceChannel, token)

	err := m.SendCoinsFromAccountToModule(ctx, sender, transfertypes.ModuleName, sdk.Coins{token})
	if err != nil {
		return err
	}

	return args.Error(0)
}
