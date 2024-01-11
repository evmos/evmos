package integration_test_util

//goland:noinspection SpellCheckingInspection
import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
	rpctypes "github.com/evmos/evmos/v16/rpc/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	inflationtypes "github.com/evmos/evmos/v16/x/inflation/v1/types"
	"math"
)

// TxSend sends amount of base coin from one to another.
func (suite *ChainIntegrationTestSuite) TxSend(from, to *itutiltypes.TestAccount, amount float64) error {
	_, _, err := suite.DeliverTx(suite.CurrentContext, from, nil, suite.buildBankSendMsg(from, to, amount))
	return err
}

// TxSendAsync is the same as TxSend but with Async delivery mode.
func (suite *ChainIntegrationTestSuite) TxSendAsync(from, to *itutiltypes.TestAccount, amount float64) error {
	_, err := suite.DeliverTxAsync(suite.CurrentContext, from, nil, suite.buildBankSendMsg(from, to, amount))
	return err
}

// buildBankSendMsg returns a MsgSend with the given parameters.
func (suite *ChainIntegrationTestSuite) buildBankSendMsg(from, to *itutiltypes.TestAccount, amount float64) *banktypes.MsgSend {
	suite.Require().NotNil(from)
	suite.Require().NotNil(to)
	suite.Require().NotZero(amount)

	return &banktypes.MsgSend{
		FromAddress: from.GetCosmosAddress().String(),
		ToAddress:   to.GetCosmosAddress().String(),
		Amount: sdk.Coins{
			sdk.Coin{
				Amount: sdk.NewInt(int64(amount * math.Pow10(18))),
				Denom:  suite.ChainConstantsConfig.GetMinDenom(),
			},
		},
	}
}

// TxSendViaEVM sends amount of base coin from one to another, via EVM module
func (suite *ChainIntegrationTestSuite) TxSendViaEVM(from, to *itutiltypes.TestAccount, amount float64) (*evmtypes.MsgEthereumTx, error) {
	msgEthereumTx := suite.buildMsgEthereumTxTransfer(from, to, amount)
	_, err := suite.DeliverEthTx(from, msgEthereumTx)
	return msgEthereumTx, err
}

// TxSendViaEVMAsync is the same as TxSendViaEVM but with Async delivery mode.
func (suite *ChainIntegrationTestSuite) TxSendViaEVMAsync(from, to *itutiltypes.TestAccount, amount float64) (*evmtypes.MsgEthereumTx, error) {
	msgEthereumTx := suite.buildMsgEthereumTxTransfer(from, to, amount)
	return msgEthereumTx, suite.DeliverEthTxAsync(from, msgEthereumTx)
}

// buildMsgEthereumTxTransfer returns a MsgEthereumTx with the given parameters.
func (suite *ChainIntegrationTestSuite) buildMsgEthereumTxTransfer(from, to *itutiltypes.TestAccount, amount float64) *evmtypes.MsgEthereumTx {
	suite.Require().NotNil(from)
	suite.Require().NotNil(to)
	suite.Require().NotZero(amount)

	toEvmAddr := to.GetEthAddress()
	amountInt := sdk.NewInt(int64(amount * math.Pow10(18)))
	return suite.prepareMsgEthereumTx(suite.CurrentContext, from, &toEvmAddr, amountInt.BigInt(), nil, 21000)
}

// QueryBalance returns the coin-base balance of given address at given context block.
// The data is read from query client.
func (suite *ChainIntegrationTestSuite) QueryBalance(height int64, cosmosAddress string) *sdk.Coin {
	return suite.QueryBalanceByDenom(height, cosmosAddress, suite.ChainConstantsConfig.GetMinDenom())
}

// QueryBalanceByDenom returns the balance of specified denom of given address at given context block.
// The data is read from query client.
func (suite *ChainIntegrationTestSuite) QueryBalanceByDenom(height int64, cosmosAddress, baseDenom string) *sdk.Coin {
	res, err := suite.QueryClientsAt(height).Bank.Balance(
		rpctypes.ContextWithHeight(height),
		&banktypes.QueryBalanceRequest{
			Address: cosmosAddress,
			Denom:   baseDenom,
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	return res.Balance
}

// QueryBalanceFromStore returns the coin-base balance of given address at given context block.
// The data is read directly from store.
func (suite *ChainIntegrationTestSuite) QueryBalanceFromStore(height int64, address sdk.AccAddress) *sdk.Coin {
	return suite.QueryBalanceByDenomFromStore(height, address, suite.ChainConstantsConfig.GetMinDenom())
}

// QueryBalanceByDenomFromStore returns the coin-base balance of a specific denom of given address at given context block.
// The data is read directly from store.
func (suite *ChainIntegrationTestSuite) QueryBalanceByDenomFromStore(height int64, address sdk.AccAddress, baseDenom string) *sdk.Coin {
	coin := suite.ChainApp.BankKeeper().GetBalance(suite.ContextAt(height), address, baseDenom)
	return &coin
}

// MintCoin mints a new amount of coin into given account.
func (suite *ChainIntegrationTestSuite) MintCoin(receiver *itutiltypes.TestAccount, coin sdk.Coin) {
	suite.Require().NotNil(receiver)

	suite.MintCoinToCosmosAddress(receiver.GetCosmosAddress(), coin)
}

// MintCoinToCosmosAddress mints a new amount of coin into given account.
func (suite *ChainIntegrationTestSuite) MintCoinToCosmosAddress(receiver sdk.AccAddress, coin sdk.Coin) {
	suite.Require().NotEmpty(receiver)

	coins := sdk.NewCoins(coin)

	err := suite.ChainApp.BankKeeper().MintCoins(suite.CurrentContext, inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)

	err = suite.ChainApp.BankKeeper().SendCoinsFromModuleToAccount(suite.CurrentContext, inflationtypes.ModuleName, receiver, coins)
	suite.Require().NoError(err)
}

// MintCoinToModuleAccount mints a new amount of coin into given module account.
func (suite *ChainIntegrationTestSuite) MintCoinToModuleAccount(receiver authtypes.ModuleAccountI, coin sdk.Coin) {
	suite.Require().NotNil(receiver)

	coins := sdk.NewCoins(coin)

	err := suite.ChainApp.BankKeeper().MintCoins(suite.CurrentContext, inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)

	err = suite.ChainApp.BankKeeper().SendCoinsFromModuleToModule(suite.CurrentContext, inflationtypes.ModuleName, receiver.GetName(), coins)
	suite.Require().NoError(err)
}

// NewBaseCoin returns an instance of sdk.Coin of base coin with given amount.
func (suite *ChainIntegrationTestSuite) NewBaseCoin(amount int64) sdk.Coin {
	intAmt := sdkmath.NewInt(amount).Mul(sdkmath.NewInt(int64(math.Pow10(18))))
	return sdk.NewCoin(suite.ChainConstantsConfig.GetMinDenom(), intAmt)
}
