package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v11/x/erc20/types"
	"github.com/evmos/evmos/v11/x/evm/statedb"
	evm "github.com/evmos/evmos/v11/x/evm/types"
	"github.com/stretchr/testify/mock"
)

var _ types.EVMKeeper = &MockEVMKeeper{}

type MockEVMKeeper struct {
	mock.Mock
}

func (m *MockEVMKeeper) GetParams(ctx sdk.Context) evm.Params {
	args := m.Called(mock.Anything)
	return args.Get(0).(evm.Params)
}

func (m *MockEVMKeeper) GetAccountWithoutBalance(ctx sdk.Context, addr common.Address) *statedb.Account {
	args := m.Called(mock.Anything, mock.Anything)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*statedb.Account)
}

func (m *MockEVMKeeper) EstimateGas(c context.Context, req *evm.EthCallRequest) (*evm.EstimateGasResponse, error) {
	args := m.Called(mock.Anything, mock.Anything)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*evm.EstimateGasResponse), args.Error(1)
}

func (m *MockEVMKeeper) ApplyMessage(ctx sdk.Context, msg core.Message, tracer vm.EVMLogger, commit bool) (*evm.MsgEthereumTxResponse, error) {
	args := m.Called(mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*evm.MsgEthereumTxResponse), args.Error(1)
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
