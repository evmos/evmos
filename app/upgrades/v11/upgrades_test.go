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
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/evmos/evmos/v10/app"
	v11 "github.com/evmos/evmos/v10/app/upgrades/v11"
	"github.com/evmos/evmos/v10/testutil"
	"github.com/stretchr/testify/suite"

	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmostypes "github.com/evmos/evmos/v10/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
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
	suite.app = app.Setup(checkTx, feemarkettypes.DefaultGenesisState())
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

func (suite *UpgradeTestSuite) setupEscrowAccounts(accCount int) {
	for i := 0; i <= accCount; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		addr := ibctypes.GetEscrowAddress(ibctypes.PortID, channelID)

		// set accounts as BaseAccounts
		baseAcc := authtypes.NewBaseAccountWithAddress(addr)
		suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)
	}
}

func (suite *UpgradeTestSuite) setValidators(validatorsAddr []string) {
	for _, a := range validatorsAddr {
		// Set Validator
		valAddr, err := sdk.ValAddressFromBech32(a)
		suite.Require().NoError(err)
		validator, err := stakingtypes.NewValidator(valAddr, suite.consKey, stakingtypes.Description{})
		suite.Require().NoError(err)
		validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
		suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
		err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
		suite.Require().NoError(err)
	}
	validators := suite.app.StakingKeeper.GetValidators(suite.ctx, 1000)
	suite.Require().Equal(len(validatorsAddr)+1, len(validators))
}

func (suite *UpgradeTestSuite) TestMigrateEscrowAcc() {
	suite.SetupTest(evmostypes.MainnetChainID)

	// fund some escrow accounts
	existingAccounts := 30
	suite.setupEscrowAccounts(existingAccounts)

	// Run migrations
	v11.MigrateEscrowAccounts(suite.ctx, suite.app.AccountKeeper)

	// check account types for channels 0 to 36
	for i := 0; i <= 36; i++ {
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
	}
}

func (suite *UpgradeTestSuite) TestDistributeRewards() {
	// define constants
	const mainnetChainID = evmostypes.MainnetChainID + "-4"
	communityPoolAccountAddress := sdk.MustAccAddressFromBech32("evmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8974jnh")

	// checks on reward amounts
	balance, ok := sdk.NewIntFromString("7399998994000000000000000")
	suite.Require().True(ok, "error converting rewards account balance")

	expRewards, ok := sdk.NewIntFromString("5625000000302600187543552")
	suite.Require().True(ok, "error converting rewards")

	actualRewards := math.NewInt(0)
	for _, currentElem := range v11.Allocations {
		res, _ := sdk.NewIntFromString(currentElem[1])
		actualRewards = actualRewards.Add(res)
	}
	suite.Require().Equal(expRewards, actualRewards)

	// get unique validator addresses in allocations
	var validatorAddresses []string
	for _, allocation := range v11.Allocations {
		if !slices.Contains(validatorAddresses, allocation[2]) {
			validatorAddresses = append(validatorAddresses, allocation[2])
		}
	}

	// create map of validator addresses to expected delegations (have to sum)
	validatorDelegations := make(map[string]sdk.Int)
	for _, allocation := range v11.Allocations {
		tokenAmount, ok := sdk.NewIntFromString(allocation[1])
		suite.Require().True(ok, "error converting token allocations")

		totalTokens, ok := validatorDelegations[allocation[2]]
		if !ok {
			validatorDelegations[allocation[2]] = tokenAmount
		} else {
			validatorDelegations[allocation[2]] = totalTokens.Add(tokenAmount)
		}
	}

	var (
		expCommPoolBalance = balance.Sub(expRewards)
		noRewardAddr       = sdk.MustAccAddressFromBech32("evmos1009egsf8sk3puq3aynt8eymmcqnneezkkvceav")
	)

	testCases := []struct {
		name            string
		chainID         string
		malleate        func()
		expectedSuccess bool
	}{
		{
			"Mainnet - success",
			mainnetChainID,
			func() {},
			true,
		},
		{
			"Mainnet - insufficient funds on reward account - fail",
			mainnetChainID,
			func() {
				suite.app.BankKeeper.SendCoins(
					suite.ctx,
					sdk.MustAccAddressFromBech32(v11.FundingAccount),
					noRewardAddr,
					sdk.NewCoins(
						sdk.NewCoin(evmostypes.BaseDenom, balance.Quo(math.NewInt(2))),
					),
				)
			},
			false,
		},
		{
			"Mainnet - invalid reward amount - fail",
			mainnetChainID,
			func() {
				v11.Allocations[0][1] = "a0151as2021231a"
			},
			false,
		},
		{
			"Testnet - no-op",
			evmostypes.TestnetChainID + "-4",
			func() {},
			false,
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
			initialDel := suite.getDelegatedTokens(validatorAddresses)
			suite.Require().Equal(math.NewInt(0), initialDel)

			if evmostypes.IsMainnet(tc.chainID) {
				err := v11.DistributeRewards(suite.ctx, suite.app.BankKeeper, suite.app.StakingKeeper, suite.app.DistrKeeper)
				if !tc.expectedSuccess {
					suite.Require().Error(err)
					return
				}
				suite.Require().NoError(err)
			}

			if tc.expectedSuccess {

				// do allocations
				for i := range v11.Allocations {
					addr := sdk.MustAccAddressFromBech32(v11.Allocations[i][0])
					valShare, _ := sdk.NewIntFromString(v11.Allocations[i][1])

					balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, evmostypes.BaseDenom)
					suite.Require().Equal(math.NewInt(0), balance.Amount)

					// get staked (delegated) tokens
					d := suite.app.StakingKeeper.GetAllDelegatorDelegations(suite.ctx, addr)

					// sum of all delegations should be equal to rewards
					delegatedAmt := suite.sumDelegations(d)
					suite.Require().Equal(valShare, delegatedAmt)
				}

				// account not in list should NOT get rewards
				// balance should be 0
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, noRewardAddr, evmostypes.BaseDenom)
				suite.Require().Equal(sdk.NewInt(0), balance.Amount)

				// get staked (delegated) tokens - no delegations expected
				d := suite.app.StakingKeeper.GetAllDelegatorDelegations(suite.ctx, noRewardAddr)
				suite.Require().Empty(d)

				// check delegation for each validator
				totalDelegations := math.NewInt(0)
				for _, v := range validatorAddresses {
					delTokens := suite.getDelegatedTokens([]string{v})

					// amount delegated should be equal to sums calculated pre-tests
					suite.Require().Equal(validatorDelegations[v], delTokens)

					totalDelegations = totalDelegations.Add(delTokens)
				}

				// sum of all delegations should be equal to rewards
				suite.Require().Equal(expRewards, totalDelegations)

				// check community pool balance
				commPoolFinalBalance := suite.app.BankKeeper.GetBalance(suite.ctx, communityPoolAccountAddress, evmostypes.BaseDenom)
				suite.Require().Equal(expCommPoolBalance, commPoolFinalBalance.Amount)

				// Funding acc balance should be 0 after the rewards distribution
				finalFundingAccBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(v11.FundingAccount), evmostypes.BaseDenom)
				suite.Require().Equal(math.NewInt(0), finalFundingAccBalance.Amount)

			} else { // no-op

				for i := range v11.Allocations {
					addr := sdk.MustAccAddressFromBech32(v11.Allocations[i][0])
					balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, evmostypes.BaseDenom)
					suite.Require().Equal(sdk.NewInt(0), balance.Amount)

					// get staked (delegated) tokens
					d := suite.app.StakingKeeper.GetAllDelegatorDelegations(suite.ctx, addr)
					suite.Require().Empty(d)
				}

				// check delegation for validators
				delTokens := suite.getDelegatedTokens(validatorAddresses)
				suite.Require().Equal(math.NewInt(0), delTokens)

				// check community pool balance
				commPoolFinalBalance := suite.app.BankKeeper.GetBalance(suite.ctx, communityPoolAccountAddress, evmostypes.BaseDenom)

				suite.Require().Equal(sdk.NewInt(0), commPoolFinalBalance.Amount)
			}
		})
	}
}

func (suite *UpgradeTestSuite) fundTestnetRewardsAcc(amount math.Int) {
	rewardsAcc, err := sdk.AccAddressFromBech32(v11.FundingAccount)
	suite.Require().NoError(err)

	rewards := sdk.NewCoins(sdk.NewCoin(evmostypes.BaseDenom, amount))
	err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, rewardsAcc, rewards)
	suite.Require().NoError(err)
}

func (suite *UpgradeTestSuite) sumDelegations(ds []stakingtypes.Delegation) math.Int {
	sum := math.NewInt(0)
	for _, d := range ds {
		val, ok := suite.app.StakingKeeper.GetValidator(suite.ctx, d.GetValidatorAddr())
		suite.Require().True(ok)
		amt := val.TokensFromShares(d.GetShares())
		sum = sum.Add(amt.TruncateInt())
	}
	return sum
}

func (suite *UpgradeTestSuite) getDelegatedTokens(valAddrs []string) math.Int {
	delTokens := math.NewInt(0)
	for _, v := range valAddrs {
		addr, err := sdk.ValAddressFromBech32(v)
		suite.Require().NoError(err)
		// get staked (delegated) tokens
		d := suite.app.StakingKeeper.GetValidatorDelegations(suite.ctx, addr)

		delegatedAmt := suite.sumDelegations(d)
		delTokens = delTokens.Add(delegatedAmt)
	}
	return delTokens
}
