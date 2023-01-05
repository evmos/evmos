package v11_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/evmos/evmos/v10/app"
	v11 "github.com/evmos/evmos/v10/app/upgrades/v11"
	"github.com/evmos/evmos/v10/testutil"
	"github.com/stretchr/testify/suite"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmostypes "github.com/evmos/evmos/v10/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.Evmos
	consAddress sdk.ConsAddress
}

func (suite *UpgradeTestSuite) SetupTest(chainID string) {
	checkTx := false

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// NOTE: this is the new binary, not the old one.
	suite.app = app.Setup(checkTx, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         chainID,
		Time:            time.Date(2022, 5, 9, 8, 0, 0, 0, time.UTC),
		ProposerAddress: suite.consAddress.Bytes(),

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
	balance, ok := sdk.NewIntFromString("7399998994000000000000000")
	suite.Require().True(ok, "error converting rewards account balance")

	rewards, ok := sdk.NewIntFromString("5625000000000000000000000")
	suite.Require().True(ok, "error converting rewards")

	var (
		valCount           = int64(len(v11.Validators))
		expDelegation      = rewards.Quo(math.NewInt(valCount))
		expCommPoolBalance = balance.Sub(rewards)
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
			evmostypes.MainnetChainID + "-4",
			func() {},
			true,
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
			suite.SetupTest(evmostypes.MainnetChainID)
			suite.fundTestnetRewardsAcc(balance)

			err := v11.DistributeRewards(suite.ctx, suite.app.BankKeeper, suite.app.StakingKeeper)
			suite.Require().NoError(err)

			if tc.expectedSuccess {
				for i := range v11.Accounts {
					addr := sdk.MustAccAddressFromBech32(v11.Accounts[i][0])
					res, _ := sdk.NewIntFromString(v11.Accounts[i][1])

					// balance should be 0 - all reward tokens are delegated
					balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, evmostypes.BaseDenom)
					suite.Require().Equal(balance.Amount, sdk.NewInt(0))

					// get staked (delegated) tokens
					d := suite.app.StakingKeeper.GetAllDelegatorDelegations(suite.ctx, addr)

					// sum of all delegations should be equal to rewards
					delegatedAmt := suite.sumDelegations(d)
					suite.Require().Equal(res, delegatedAmt)
				}

				// account not in list should NOT get rewards
				// balance should be 0 - all reward tokens are delegated
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, noRewardAddr, evmostypes.BaseDenom)
				suite.Require().Equal(balance.Amount, sdk.NewInt(0))

				// get staked (delegated) tokens
				d := suite.app.StakingKeeper.GetAllDelegatorDelegations(suite.ctx, noRewardAddr)
				suite.Require().Empty(d)

				// check delegation for each validator
				for _, v := range v11.Validators {
					addr, err := sdk.ValAddressFromBech32(v)
					suite.Require().NoError(err)
					// get staked (delegated) tokens
					d := suite.app.StakingKeeper.GetValidatorDelegations(suite.ctx, addr)

					// sum of all delegations should be equal to rewards
					delegatedAmt := suite.sumDelegations(d)
					suite.Require().Equal(expDelegation, delegatedAmt)
				}
				// check community pool balance
				commPoolFinalBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(v11.CommunityPoolAccount), evmostypes.BaseDenom)

				suite.Require().Equal(expCommPoolBalance, commPoolFinalBalance.Amount)
			} else {
				for i := range v11.Accounts {
					addr := sdk.MustAccAddressFromBech32(v11.Accounts[i][0])
					balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, evmostypes.BaseDenom)
					suite.Require().Equal(sdk.NewInt(0), balance.Amount)

					// get staked (delegated) tokens
					d := suite.app.StakingKeeper.GetAllDelegatorDelegations(suite.ctx, addr)
					suite.Require().Empty(d)
				}
				// check delegation for each validator
				for _, v := range v11.Validators {
					addr, err := sdk.ValAddressFromBech32(v)
					suite.Require().NoError(err)
					// get staked (delegated) tokens
					d := suite.app.StakingKeeper.GetValidatorDelegations(suite.ctx, addr)
					suite.Require().Empty(d)
				}

				// check community pool balance
				commPoolFinalBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(v11.CommunityPoolAccount), evmostypes.BaseDenom)

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
	var sum math.Int
	for _, d := range ds {
		val, ok := suite.app.StakingKeeper.GetValidator(suite.ctx, d.GetValidatorAddr())
		suite.Require().True(ok)
		amt := val.TokensFromShares(d.GetShares())
		sum.Add(amt.RoundInt())
	}
	return sum
}
