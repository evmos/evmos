package keeper_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	"github.com/evmos/evmos/v11/x/recovery/types"
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

func (m *MockTransferKeeper) Transfer(goCtx context.Context, msg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error) {
	args := m.Called(
		mock.Anything,
		mock.Anything,
	)
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = m.SendCoinsFromAccountToModule(ctx, sender, transfertypes.ModuleName, sdk.Coins{msg.Token})
	if err != nil {
		return nil, err
	}

	return nil, args.Error(1)
}
