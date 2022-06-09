package v5_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/tharsis/evmos/v5/app"
	v5 "github.com/tharsis/evmos/v5/app/upgrades/v5"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.Evmos
	consAddress sdk.ConsAddress
}

func (suite *UpgradeTestSuite) SetupTest() {
	checkTx := false

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// NOTE: this is the new binary, not the old one.
	suite.app = app.Setup(checkTx, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9001-2",
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
}

func TestUpgradeTestSuite(t *testing.T) {
	s := new(UpgradeTestSuite)
	suite.Run(t, s)
}

func (suite *UpgradeTestSuite) TestScheduledUpgrade() {
	testCases := []struct {
		name       string
		preUpdate  func()
		update     func()
		postUpdate func()
	}{
		{
			"scheduled upgrade",
			func() {
				plan := types.Plan{
					Name:   v5.UpgradeName,
					Height: v5.MainnetUpgradeHeight,
					Info:   v5.UpgradeInfo,
				}
				err := suite.app.UpgradeKeeper.ScheduleUpgrade(suite.ctx, plan)
				suite.Require().NoError(err)

				// ensure the plan is scheduled
				plan, found := suite.app.UpgradeKeeper.GetUpgradePlan(suite.ctx)
				suite.Require().True(found)
			},
			func() {
				suite.ctx = suite.ctx.WithBlockHeight(v5.MainnetUpgradeHeight)
				suite.Require().NotPanics(
					func() {
						beginBlockRequest := abci.RequestBeginBlock{
							Header: suite.ctx.BlockHeader(),
						}
						suite.app.BeginBlocker(suite.ctx, beginBlockRequest)
					},
				)
			},
			func() {
				// check that the default params have been overridden by the init function
				fmParams := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				suite.Require().Equal(sdk.NewDecWithPrec(25, 3).String(), fmParams.MinGasPrice.String())
				suite.Require().Equal(sdk.NewDecWithPrec(5, 1).String(), fmParams.MinGasMultiplier.String())
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.preUpdate()
			tc.update()
			tc.postUpdate()
		})
	}
}
