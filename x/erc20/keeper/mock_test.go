package keeper_test

import (
	"context"

	"github.com/stretchr/testify/mock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v11/x/erc20/types"
	"github.com/evmos/evmos/v11/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"
)

var _ types.EVMKeeper = &MockEVMKeeper{}

type MockEVMKeeper struct {
	mock.Mock
}

func (m *MockEVMKeeper) GetParams(ctx sdk.Context) evmtypes.Params {
	args := m.Called(mock.Anything)
	return args.Get(0).(evmtypes.Params)
}

func (m *MockEVMKeeper) GetAccountWithoutBalance(ctx sdk.Context, addr common.Address) *statedb.Account {
	args := m.Called(mock.Anything, mock.Anything)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*statedb.Account)
}

func (m *MockEVMKeeper) EstimateGas(c context.Context, req *evmtypes.EthCallRequest) (*evmtypes.EstimateGasResponse, error) {
	args := m.Called(mock.Anything, mock.Anything)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*evmtypes.EstimateGasResponse), args.Error(1)
}

func (m *MockEVMKeeper) CallEVMWithData(
	ctx sdk.Context,
	from common.Address,
	contract *common.Address,
	data []byte,
	commit bool,
) (*evmtypes.MsgEthereumTxResponse, error) {
	args := m.Called(mock.Anything)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*evmtypes.MsgEthereumTxResponse), args.Error(1)
}

func (m *MockEVMKeeper) CallEVM(
	ctx sdk.Context,
	abi abi.ABI,
	from, contract common.Address,
	commit bool,
	method string,
	args ...interface{},
) (*evmtypes.MsgEthereumTxResponse, error) {
	mockArgs := m.Called(mock.Anything)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}

	return mockArgs.Get(0).(*evmtypes.MsgEthereumTxResponse), mockArgs.Error(1)
}

var _ types.BankKeeper = &MockBankKeeper{}

type MockBankKeeper struct {
	mock.Mock
}

func (b *MockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	args := b.Called(mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	return args.Error(0)
}

func (b *MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	args := b.Called(mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	return args.Error(0)
}

func (b *MockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	args := b.Called(mock.Anything, mock.Anything, mock.Anything)
	return args.Error(0)
}

func (b *MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	args := b.Called(mock.Anything, mock.Anything, mock.Anything)
	return args.Error(0)
}

func (b *MockBankKeeper) IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool {
	args := b.Called(mock.Anything, mock.Anything)
	return args.Bool(0)
}

func (b *MockBankKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	args := b.Called(mock.Anything)
	return args.Bool(0)
}

func (b *MockBankKeeper) GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool) {
	args := b.Called(mock.Anything, mock.Anything)
	return args.Get(0).(banktypes.Metadata), args.Bool(1)
}

func (b *MockBankKeeper) SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata) {
}

func (b *MockBankKeeper) HasSupply(ctx sdk.Context, denom string) bool {
	args := b.Called(mock.Anything, mock.Anything)
	return args.Bool(0)
}

func (b *MockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	args := b.Called(mock.Anything, mock.Anything)
	return args.Get(0).(sdk.Coin)
}
