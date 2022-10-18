package keeper_test

import (
	"github.com/stretchr/testify/mock"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"

	"github.com/evmos/evmos/v9/x/recovery/types"
)

var _ types.TransferKeeper = &MockTransferKeeper{}

// MockTransferKeeper defines a mocked object that implements the TransferKeeper
// interface. It's used on tests to abstract the complexity of IBC transfers.
// NOTE: Bank keeper logic is not mocked since we want to test that balance has
// been updated for sender and recipient.
type MockTransferKeeper struct {
	mock.Mock
	bankkeeper.Keeper
}

func (m *MockTransferKeeper) GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (transfertypes.DenomTrace, bool) {
	args := m.Called(mock.Anything, denomTraceHash)
	return args.Get(0).(transfertypes.DenomTrace), args.Bool(1)
}

func (m *MockTransferKeeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	args := m.Called(mock.Anything, sourcePort, sourceChannel, token, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	err := m.SendCoinsFromAccountToModule(ctx, sender, transfertypes.ModuleName, sdk.Coins{token})
	if err != nil {
		return err
	}

	return args.Error(0)
}
