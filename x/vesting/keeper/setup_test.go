package keeper_test

import (
	"errors"
	"math"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/suite"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/evmos/evmos/v18/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v18/encoding"
	"github.com/evmos/evmos/v18/testutil"
	utiltx "github.com/evmos/evmos/v18/testutil/tx"
	evmostypes "github.com/evmos/evmos/v18/types"
	"github.com/evmos/evmos/v18/utils"
	epochstypes "github.com/evmos/evmos/v18/x/epochs/types"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"

	"github.com/evmos/evmos/v18/app"
	"github.com/evmos/evmos/v18/x/vesting/types"
)

var (
	contract  common.Address
	contract2 common.Address
)

var (
	erc20Name     = "Coin Token"
	erc20Symbol   = "CTKN"
	erc20Name2    = "Coin Token 2"
	erc20Symbol2  = "CTKN2"
	erc20Decimals = uint8(18)
)

type KeeperTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	app            *app.Evmos
	queryClientEvm evmtypes.QueryClient
	queryClient    types.QueryClient
	address        common.Address
	consAddress    sdk.ConsAddress
	validator      stakingtypes.Validator
	clientCtx      client.Context
	ethSigner      ethtypes.Signer
	priv           cryptotypes.PrivKey
	signer         keyring.Signer
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)
}

func (suite *KeeperTestSuite) SetupTest() error {
	checkTx := false

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	if err != nil {
		return err
	}
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = utiltx.NewSigner(priv)
	suite.priv = priv

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	if err != nil {
		return err
	}
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// Init app
	chainID := utils.TestnetChainID + "-1"
	suite.app = app.Setup(checkTx, nil, chainID)

	// Set Context
	header := testutil.NewHeader(
		1, time.Now().UTC(), chainID, suite.consAddress, nil, nil,
	)
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	// Setup query helpers
	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evmtypes.NewQueryClient(queryHelperEvm)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.VestingKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	// Set epoch start time and height for all epoch identifiers from the epoch
	// module
	identifiers := []string{epochstypes.WeekEpochID, epochstypes.DayEpochID}
	for _, identifier := range identifiers {
		epoch, found := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, identifier)
		if !found {
			return errors.New("epoch info not found")
		}
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
	amt := sdkmath.NewInt(int64(math.Pow10(18) * 2))
	err = testutil.FundAccount(
		suite.ctx,
		suite.app.BankKeeper,
		suite.priv.PubKey().Address().Bytes(),
		sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amt)),
	)
	if err != nil {
		return err
	}

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	if err != nil {
		return err
	}
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper.Keeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, validator.GetOperator())
	if err != nil {
		return err
	}
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	if err != nil {
		return err
	}
	validators := s.app.StakingKeeper.GetValidators(s.ctx, 1)
	suite.validator = validators[0]

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())

	// Deploy contracts
	contract, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
	if err != nil {
		return err
	}
	contract2, err = suite.DeployContract(erc20Name2, erc20Symbol2, erc20Decimals)
	if err != nil {
		return err
	}

	// Set correct denom in govKeeper
	govParams := suite.app.GovKeeper.GetParams(suite.ctx)
	govParams.MinDeposit = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, sdkmath.NewInt(1e6)))
	votingPeriod := time.Second
	govParams.VotingPeriod = &votingPeriod

	return suite.app.GovKeeper.SetParams(suite.ctx, govParams)
}
