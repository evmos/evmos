package keeper_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v13/app"
	"github.com/evmos/evmos/v13/contracts"
	"github.com/evmos/evmos/v13/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v13/encoding"
	"github.com/evmos/evmos/v13/server/config"
	"github.com/evmos/evmos/v13/testutil"
	utiltx "github.com/evmos/evmos/v13/testutil/tx"
	evmostypes "github.com/evmos/evmos/v13/types"
	"github.com/evmos/evmos/v13/utils"
	epochstypes "github.com/evmos/evmos/v13/x/epochs/types"
	evm "github.com/evmos/evmos/v13/x/evm/types"
	"github.com/evmos/evmos/v13/x/incentives/types"
	"github.com/stretchr/testify/require"
)

var (
	contract  common.Address
	contract2 common.Address
)

var (
	participant     = utiltx.GenerateAddress()
	participant2    = utiltx.GenerateAddress()
	denomMint       = evm.DefaultEVMDenom
	denomCoin       = "acoin"
	allocationRate  = int64(5)
	mintAllocations = sdk.DecCoins{
		sdk.NewDecCoinFromDec(denomMint, sdk.NewDecWithPrec(allocationRate, 2)),
	}
	allocations = sdk.DecCoins{
		sdk.NewDecCoinFromDec(denomMint, sdk.NewDecWithPrec(allocationRate, 2)),
		sdk.NewDecCoinFromDec(denomCoin, sdk.NewDecWithPrec(allocationRate, 2)),
	}
	epochs        = uint32(10)
	erc20Name     = "Coin Token"
	erc20Symbol   = "CTKN"
	erc20Name2    = "Coin Token 2"
	erc20Symbol2  = "CTKN2"
	erc20Decimals = uint8(18)
)

// Test helpers
func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	checkTx := false

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = utiltx.NewSigner(priv)
	suite.priv = priv

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// Init app
	suite.app = app.Setup(checkTx, nil)

	// Set Context
	header := testutil.NewHeader(
		1, time.Now().UTC(), "evmos_9001-1", suite.consAddress, nil, nil,
	)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, header)

	// Setup query helpers
	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelperEvm)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.IncentivesKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	// Set epoch start time and height for all epoch identifiers from the epoch
	// module
	identifiers := []string{epochstypes.WeekEpochID, epochstypes.DayEpochID}
	for _, identifier := range identifiers {
		epoch, found := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, identifier)
		suite.Require().True(found)
		epoch.StartTime = suite.ctx.BlockTime()
		epoch.CurrentEpochStartHeight = suite.ctx.BlockHeight()
		suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epoch)
	}

	acc := &evmostypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	// fund signer acc to pay for tx fees
	amt := sdk.NewInt(int64(math.Pow10(18) * 2))
	err = testutil.FundAccount(
		suite.ctx,
		suite.app.BankKeeper,
		suite.priv.PubKey().Address().Bytes(),
		sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amt)),
	)
	suite.Require().NoError(err)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
}

func (suite *KeeperTestSuite) deployContracts() {
	// Deploy contracts
	contract, _ = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
	contract2, _ = suite.DeployContract(erc20Name2, erc20Symbol2, erc20Decimals)
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	var err error
	suite.ctx, err = testutil.Commit(suite.ctx, suite.app, t, nil)
	suite.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelper)
}

// MintFeeCollector mints coins with the bank modules and sends them to the fee
// collector.
func (suite *KeeperTestSuite) MintFeeCollector(coins sdk.Coins) {
	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, authtypes.FeeCollectorName, coins)
	suite.Require().NoError(err)
}

// DeployContract deploys the ERC20MinterBurnerDecimalsContract.
func (suite *KeeperTestSuite) DeployContract(name, symbol string, decimals uint8) (common.Address, error) {
	suite.Commit()
	addr, err := testutil.DeployContract(
		suite.ctx,
		suite.app,
		suite.priv,
		suite.queryClientEvm,
		contracts.ERC20MinterBurnerDecimalsContract,
		name, symbol, decimals,
	)
	suite.Commit()
	return addr, err
}

// MintERC20Token mints ERC20MinterBurnerDecimalsContract tokens..
func (suite *KeeperTestSuite) MintERC20Token(
	contractAddr,
	from, to common.Address,
	amount *big.Int,
) *evm.MsgEthereumTx {
	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("mint", to, amount)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

// BurnERC20Token burns ERC20MinterBurnerDecimalsContract tokens.
func (suite *KeeperTestSuite) BurnERC20Token(
	contractAddr,
	from common.Address,
	amount *big.Int,
) *evm.MsgEthereumTx {
	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("transfer", types.ModuleAddress, amount)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

// GrantERC20Token grants ERC20MinterBurnerDecimalsContract tokens.
func (suite *KeeperTestSuite) GrantERC20Token(
	contractAddr,
	from, to common.Address,
	roleString string,
) *evm.MsgEthereumTx {
	// 0xCc508cD0818C85b8b8a1aB4cEEef8d981c8956A6 MINTER_ROLE
	role := crypto.Keccak256([]byte(roleString))
	// needs to be an array not a slice
	var v [32]byte
	copy(v[:], role)

	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("grantRole", v, to)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

// TransferERC20Token transfers tokens from one account to another to another
func (suite *KeeperTestSuite) TransferERC20Token(
	contractAddr,
	from, to common.Address,
	amount *big.Int,
) *evm.MsgEthereumTx {
	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("transfer", to, amount)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

// sendTx creates, sings and sends a evm transaction from suite.address account.
func (suite *KeeperTestSuite) sendTx(
	contractAddr,
	from common.Address,
	transferData []byte,
) *evm.MsgEthereumTx {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	args, err := json.Marshal(&evm.TransactionArgs{To: &contractAddr, From: &from, Data: (*hexutil.Bytes)(&transferData)})
	suite.Require().NoError(err)

	res, err := suite.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: config.DefaultGasCap,
	})
	suite.Require().NoError(err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	// Mint the max gas to the FeeCollector to ensure balance in case of refund
	suite.MintFeeCollector(sdk.NewCoins(sdk.NewCoin(evm.DefaultEVMDenom, sdk.NewInt(suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx).Int64()*int64(res.Gas)))))

	ethTxParams := evm.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &contractAddr,
		GasLimit:  res.Gas,
		GasFeeCap: suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		GasPrice:  big.NewInt(1),
		Input:     transferData,
		Accesses:  &ethtypes.AccessList{},
	}
	ercTransferTx := evm.NewTx(&ethTxParams)

	ercTransferTx.From = from.Hex()

	err = ercTransferTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, ercTransferTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)
	return ercTransferTx
}

// BalanceOf gets the ERC20MinterBurnerDecimalsContract token balance at a given
// addr.
func (suite *KeeperTestSuite) BalanceOf(contract, account common.Address) *big.Int {
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI

	res, err := suite.app.Erc20Keeper.CallEVM(suite.ctx, erc20, types.ModuleAddress, contract, false, "balanceOf", account)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	unpacked, err := erc20.Unpack("balanceOf", res.Ret)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(unpacked)
	suite.Require().IsType(unpacked[0], &big.Int{})

	return unpacked[0].(*big.Int)
}

// NameOf gets the name of a given ERC20MinterBurnerDecimalsContract contract.
func (suite *KeeperTestSuite) NameOf(contract common.Address) string {
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI

	res, err := suite.app.Erc20Keeper.CallEVM(suite.ctx, erc20, types.ModuleAddress, contract, false, "name")
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	unpacked, err := erc20.Unpack("name", res.Ret)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(unpacked)

	return fmt.Sprintf("%v", unpacked[0])
}

// ensureHooksSet tries to set the hooks on EVMKeeper, this will fail if the
// incentives hook is already set
func (suite *KeeperTestSuite) ensureHooksSet() {
	defer func() {
		err := recover()
		suite.Require().NotNil(err)
	}()
	suite.app.EvmKeeper.SetHooks(suite.app.IncentivesKeeper.Hooks())
}
