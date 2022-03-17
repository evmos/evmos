package keeper_test

import (
	"github.com/stretchr/testify/mock"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	return args.Error(0)
}

// // TestSomethingWithPlaceholder is a second example of how to use our test object to
// // make assertions about some target code we are testing.
// // This time using a placeholder. Placeholders might be used when the
// // data being passed in is normally dynamically generated and cannot be
// // predicted beforehand (eg. containing hashes that are time sensitive)
// func TestSomethingWithPlaceholder(t *testing.T) {

//   // create an instance of our test object
// 	mockedTransferKeeper := new(TransferKeeper)

//   // setup expectations with a placeholder in the argument list
// 	mockedTransferKeeper.On("SendTransfer", mock.Anything).Return(nil)

//   // call the code we are testing
// 	mockedTransferKeeper =

//   // assert that the expectations were met
// 	mockedTransferKeeper.AssertExpectations(t)

// }
