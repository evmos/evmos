package v7_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/evmos/evmos/v8/app"
	v7 "github.com/evmos/evmos/v8/app/upgrades/v7"
	"github.com/evmos/evmos/v8/testutil"
	evmostypes "github.com/evmos/evmos/v8/types"
	claimstypes "github.com/evmos/evmos/v8/x/claims/types"
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

func (suite *UpgradeTestSuite) TestMigrateFaucetBalance() {
	from := sdk.MustAccAddressFromBech32(v7.FaucetAddressFrom)
	to := sdk.MustAccAddressFromBech32(v7.FaucetAddressTo)

	testCases := []struct {
		name              string
		chainID           string
		expectedMigration bool
	}{
		{
			"Testnet - sucess",
			evmostypes.TestnetChainID + "-4",
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest(tc.chainID) // reset

			coins := sdk.NewCoins(sdk.NewCoin(suite.app.StakingKeeper.BondDenom(suite.ctx), sdk.NewInt(1000)))
			err := testutil.FundAccount(suite.app.BankKeeper, suite.ctx, from, coins)
			suite.Require().NoError(err)

			suite.Require().NotPanics(func() {
				v7.MigrateFaucetBalances(suite.ctx, suite.app.BankKeeper)
				suite.app.Commit()
			})

			balancesFrom := suite.app.BankKeeper.GetAllBalances(suite.ctx, from)
			balancesTo := suite.app.BankKeeper.GetAllBalances(suite.ctx, to)

			if tc.expectedMigration {
				suite.Require().True(balancesFrom.IsZero())
				suite.Require().Equal(coins, balancesTo)
			} else {
				suite.Require().Equal(coins, balancesFrom)
				suite.Require().Nil(balancesTo)
			}
		})
	}
}

func (suite *UpgradeTestSuite) TestMigrateSkippedEpochs() {
	testCases := []struct {
		name                  string
		chainID               string
		malleate              func()
		expectedSkippedEpochs uint64
	}{
		{
			"success",
			evmostypes.MainnetChainID + "-2",
			func() {
				suite.app.InflationKeeper.SetSkippedEpochs(suite.ctx, uint64(94))
			},
			uint64(92),
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest(tc.chainID) // reset

			tc.malleate()

			suite.Require().NotPanics(func() {
				v7.MigrateSkippedEpochs(suite.ctx, suite.app.InflationKeeper)
			})

			newSkippedEpochs := suite.app.InflationKeeper.GetSkippedEpochs(suite.ctx)
			suite.Require().Equal(tc.expectedSkippedEpochs, newSkippedEpochs)
		})
	}
}

func (suite *UpgradeTestSuite) TestMigrateClaim() {
	from, err := sdk.AccAddressFromBech32(v7.ContributorAddrFrom)
	suite.Require().NoError(err)
	to, err := sdk.AccAddressFromBech32(v7.ContributorAddrTo)
	suite.Require().NoError(err)
	cr := claimstypes.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(100), ActionsCompleted: []bool{false, false, false, false}}

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"with claims record",
			func() {
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, from, cr)
			},
			true,
		},
		{
			"without claims record",
			func() {
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest(evmostypes.TestnetChainID + "-2") // reset

			tc.malleate()

			v7.MigrateContributorClaim(suite.ctx, suite.app.ClaimsKeeper)

			_, foundFrom := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, from)
			crTo, foundTo := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, to)
			if tc.expPass {
				suite.Require().False(foundFrom)
				suite.Require().True(foundTo)
				suite.Require().Equal(crTo, cr)
			} else {
				suite.Require().False(foundTo)
			}
		})
	}
}
