package v11_test

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/exp/slices"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/evmos/evmos/v16/app"
	v11 "github.com/evmos/evmos/v16/app/upgrades/v11"
	"github.com/evmos/evmos/v16/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v16/testutil"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	feemarkettypes "github.com/evmos/evmos/v16/x/feemarket/types"
	"github.com/stretchr/testify/suite"

	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v16/utils"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	app     *app.Evmos
	consKey cryptotypes.PubKey
}

func (suite *UpgradeTestSuite) SetupTest(chainID string) {
	checkTx := false

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	suite.consKey = priv.PubKey()
	consAddress := sdk.ConsAddress(priv.PubKey().Address())

	// NOTE: this is the new binary, not the old one.
	suite.app = app.Setup(checkTx, feemarkettypes.DefaultGenesisState(), chainID)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         chainID,
		Time:            time.Date(2022, 5, 9, 8, 0, 0, 0, time.UTC),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	cp := suite.app.BaseApp.GetConsensusParams(suite.ctx)
	suite.ctx = suite.ctx.WithConsensusParams(cp)
}

func TestUpgradeTestSuite(t *testing.T) {
	s := new(UpgradeTestSuite)
	suite.Run(t, s)
}

var expAccNum = make(map[int]uint64)

func (suite *UpgradeTestSuite) setupEscrowAccounts(accCount int) {
	for i := 0; i <= accCount; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		addr := ibctypes.GetEscrowAddress(ibctypes.PortID, channelID)

		// set accounts as BaseAccounts
		baseAcc := authtypes.NewBaseAccountWithAddress(addr)
		err := baseAcc.SetAccountNumber(suite.app.AccountKeeper.NextAccountNumber(suite.ctx))
		suite.Require().NoError(err)
		expAccNum[i] = baseAcc.AccountNumber
		suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)
	}
}

func (suite *UpgradeTestSuite) setValidators(validatorsAddr []string) {
	for _, valAddrStr := range validatorsAddr {
		// Set Validator
		valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
		suite.Require().NoError(err)

		validator, err := stakingtypes.NewValidator(valAddr, suite.consKey, stakingtypes.Description{})
		suite.Require().NoError(err)

		validator = stakingkeeper.TestingUpdateValidator(&suite.app.StakingKeeper, suite.ctx, validator, true)

		err = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, validator.GetOperator())
		suite.Require().NoError(err)
		err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
		suite.Require().NoError(err)
	}

	validators := suite.app.StakingKeeper.GetValidators(suite.ctx, 1000)
	suite.Require().Equal(len(validatorsAddr)+1, len(validators))
}

func (suite *UpgradeTestSuite) TestMigrateEscrowAcc() {
	suite.SetupTest(utils.MainnetChainID + "-1")

	// fund some escrow accounts
	existingAccounts := 30
	suite.setupEscrowAccounts(existingAccounts)

	// Run migrations
	v11.MigrateEscrowAccounts(suite.ctx, suite.app.Logger(), suite.app.AccountKeeper)

	// check account types for channels 0 to 37
	for i := 0; i <= v11.OpenChannels; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		addr := ibctypes.GetEscrowAddress(ibctypes.PortID, channelID)
		acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr)

		if i > existingAccounts {
			suite.Require().Nil(acc, "This account did not exist, it should not be migrated")
			continue
		}
		suite.Require().NotNil(acc)

		moduleAcc, isModuleAccount := acc.(*authtypes.ModuleAccount)
		suite.Require().True(isModuleAccount)
		suite.Require().NoError(moduleAcc.Validate(), "account validation failed")

		// Check account number
		suite.Require().Equal(expAccNum[i], moduleAcc.GetAccountNumber())
	}
}

func (suite *UpgradeTestSuite) TestDistributeRewards() {
	// define constants
	mainnetChainID := utils.MainnetChainID + "-4"
	communityPool := sdk.MustAccAddressFromBech32("evmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8974jnh")
	fundingAcc := sdk.MustAccAddressFromBech32(v11.FundingAccount)

	// checks on reward amounts
	balance, ok := math.NewIntFromString("7399998994000000000000000")
	suite.Require().True(ok, "error converting rewards account balance")

	expRewards, ok := math.NewIntFromString("5624999999983399933050880")
	suite.Require().True(ok, "error converting rewards")

	var validatorAddresses []string
	validatorDelegations := make(map[string]math.Int)
	actualRewards := math.ZeroInt()

	for _, allocation := range v11.Allocations {
		amt, ok := math.NewIntFromString(allocation[1])
		suite.Require().True(ok, "failed to convert allocation")

		actualRewards = actualRewards.Add(amt)

		if !slices.Contains(validatorAddresses, allocation[2]) {
			validatorAddresses = append(validatorAddresses, allocation[2])
		}

		totalTokens, ok := validatorDelegations[allocation[2]]
		if !ok {
			validatorDelegations[allocation[2]] = amt
		} else {
			validatorDelegations[allocation[2]] = totalTokens.Add(amt)
		}
	}

	suite.Require().Equal(expRewards, actualRewards)

	expCommPoolBalance := balance.Sub(expRewards)
	noRewardAddr := sdk.MustAccAddressFromBech32("evmos1009egsf8sk3puq3aynt8eymmcqnneezkkvceav")

	testCases := []struct {
		name               string
		chainID            string
		malleate           func()
		expectedSuccess    bool
		expCommPoolBalance math.Int
	}{
		{
			"Mainnet - success",
			mainnetChainID,
			func() {},
			true,
			expCommPoolBalance,
		},
		{
			"Mainnet - insufficient funds on reward account - fail",
			mainnetChainID,
			func() {
				err := suite.app.BankKeeper.SendCoins(
					suite.ctx,
					fundingAcc,
					sdk.AccAddress(utiltx.GenerateAddress().Bytes()),
					sdk.NewCoins(
						sdk.NewCoin(utils.BaseDenom, balance.Quo(math.NewInt(2))),
					),
				)
				suite.NoError(err)
			},
			false,
			math.ZeroInt(),
		},
		{
			"Mainnet - invalid reward amount - fail",
			mainnetChainID,
			func() {
				v11.Allocations[0][1] = "a0151as2021231a"
			},
			false,
			math.ZeroInt(),
		},
		{
			"Testnet - no-op",
			utils.TestnetChainID + "-4",
			func() {},
			false,
			math.ZeroInt(),
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest(tc.chainID)
			suite.fundTestnetRewardsAcc(balance)
			tc.malleate()

			// create validators
			suite.setValidators(validatorAddresses)

			// check no delegations for validators initially
			initialDel := suite.getDelegatedTokens(validatorAddresses...)
			suite.Require().Equal(math.ZeroInt(), initialDel)

			if utils.IsMainnet(tc.chainID) {
				v11.HandleRewardDistribution(suite.ctx, suite.app.Logger(), suite.app.BankKeeper, suite.app.StakingKeeper, suite.app.DistrKeeper)
			}

			// account not in list should NOT get rewards
			// balance should be 0
			balance := suite.app.BankKeeper.GetBalance(suite.ctx, noRewardAddr, utils.BaseDenom)
			suite.Require().Equal(math.ZeroInt(), balance.Amount)

			// get staked (delegated) tokens - no delegations expected
			delegated := suite.app.StakingKeeper.GetAllDelegatorDelegations(suite.ctx, noRewardAddr)
			suite.Require().Empty(delegated)

			commPoolFinalBalance := suite.app.BankKeeper.GetBalance(suite.ctx, communityPool, utils.BaseDenom)
			suite.Require().Equal(tc.expCommPoolBalance, commPoolFinalBalance.Amount)

			// do allocations
			for i := range v11.Allocations {
				addr := sdk.MustAccAddressFromBech32(v11.Allocations[i][0])
				valShare, _ := math.NewIntFromString(v11.Allocations[i][1])

				balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, utils.BaseDenom)
				suite.Require().Equal(math.ZeroInt(), balance.Amount)

				// get staked (delegated) tokens
				delegated := suite.app.StakingKeeper.GetAllDelegatorDelegations(suite.ctx, addr)
				if tc.expectedSuccess {
					// sum of all delegations should be equal to rewards
					delegatedAmt := suite.sumDelegatorDelegations(delegated...)
					suite.Require().Equal(valShare, delegatedAmt)
				} else {
					suite.Require().Empty(delegated)
				}
			}

			// check delegation for each validator
			totalDelegations := math.ZeroInt()
			for _, v := range validatorAddresses {
				delTokens := suite.getDelegatedTokens(v)
				if tc.expectedSuccess {
					// amount delegated should be equal to sums calculated pre-tests
					suite.Require().Equal(validatorDelegations[v], delTokens)
				} else {
					suite.Require().Equal(math.ZeroInt(), delTokens)
				}
				totalDelegations = totalDelegations.Add(delTokens)
			}

			if tc.expectedSuccess {
				// sum of all delegations should be equal to rewards
				suite.Require().Equal(expRewards, totalDelegations)
				// Funding acc balance should be 0 after the rewards distribution
				finalFundingAccBalance := suite.app.BankKeeper.GetBalance(suite.ctx, fundingAcc, utils.BaseDenom)
				suite.Require().Equal(math.NewInt(0), finalFundingAccBalance.Amount)
			}
		})
	}
}

func (suite *UpgradeTestSuite) fundTestnetRewardsAcc(amount math.Int) {
	rewardsAcc, err := sdk.AccAddressFromBech32(v11.FundingAccount)
	suite.Require().NoError(err)

	rewards := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amount))
	err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, rewardsAcc, rewards)
	suite.Require().NoError(err)
}

func (suite *UpgradeTestSuite) sumDelegatorDelegations(ds ...stakingtypes.Delegation) math.Int {
	sumDec := math.LegacyNewDec(0)

	for _, d := range ds {
		validator, ok := suite.app.StakingKeeper.GetValidator(suite.ctx, d.GetValidatorAddr())
		suite.Require().True(ok)

		amt := validator.TokensFromShares(d.GetShares())
		sumDec = sumDec.Add(amt)
	}

	return sumDec.TruncateInt()
}

func (suite *UpgradeTestSuite) sumValidatorDelegations(validator stakingtypes.ValidatorI, ds ...stakingtypes.Delegation) math.Int {
	sumDec := math.LegacyNewDec(0)

	for _, d := range ds {
		amt := validator.TokensFromShares(d.GetShares())
		sumDec = sumDec.Add(amt)
	}

	return sumDec.TruncateInt()
}

func (suite *UpgradeTestSuite) getDelegatedTokens(valAddrs ...string) math.Int {
	delTokens := math.NewInt(0)

	for _, valAddrStr := range valAddrs {
		valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
		suite.Require().NoError(err)

		val, ok := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
		suite.Require().True(ok)

		// get staked (delegated) tokens
		delegations := suite.app.StakingKeeper.GetValidatorDelegations(suite.ctx, valAddr)

		delegatedAmt := suite.sumValidatorDelegations(val, delegations...)
		delTokens = delTokens.Add(delegatedAmt)
	}
	return delTokens
}

func (suite *UpgradeTestSuite) TestNoDuplicateAddress() {
	participants := make(map[string]bool)
	for _, allocation := range v11.Allocations {
		suite.Require().False(participants[allocation[0]], "duplicated allocation entry")
		participants[allocation[0]] = true
	}
}
