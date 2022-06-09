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
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/tests"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/tharsis/evmos/v5/app"
	v5 "github.com/tharsis/evmos/v5/app/upgrades/v5"
	claimskeeper "github.com/tharsis/evmos/v5/x/claims/keeper"
	claimstypes "github.com/tharsis/evmos/v5/x/claims/types"
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

func (suite *UpgradeTestSuite) TestAirdropHandle() {

	testCases := []struct {
		name     string
		original []bool
		expected []bool
	}{
		{
			"EVM-IBC claimed",
			[]bool{false, false, true, true},
			// Swap ibc<->vote
			[]bool{true, false, true, false},
		},
		{
			"DELEGATE-IBC claimed",
			[]bool{false, true, false, true},
			// Swap ibc<->evm
			[]bool{false, true, true, false},
		},
		{
			"VOTE-IBC claimed",
			[]bool{true, false, false, true},
			// Swap ibc<->evm
			[]bool{true, false, true, false},
		},
		{
			"VOTE claimed",
			[]bool{true, false, false, false},
			// Swap vote<->evm
			[]bool{false, false, true, false},
		},
		{
			"Nothing changes",
			[]bool{false, false, false, false},
			[]bool{false, false, false, false},
		},
		{
			"EVM unclaimed",
			[]bool{true, true, false, true},
			// Swap ibc<->evm
			[]bool{true, true, true, false},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			suite.ctx = suite.ctx.WithChainID("evmos_9001-1")
			addr := addClaimRecord(suite.ctx, suite.app.ClaimsKeeper, tc.original)
			vm := suite.app.UpgradeKeeper.GetModuleVersionMap(suite.ctx)

			cfg := module.NewConfigurator(suite.app.AppCodec(), suite.app.MsgServiceRouter(), suite.app.GRPCQueryRouter())

			handlerFn := v5.CreateUpgradeHandler(suite.app.ModuleManager(), cfg, suite.app.BankKeeper, suite.app.ClaimsKeeper)
			_, err := handlerFn(suite.ctx, types.Plan{}, vm)

			cr, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, addr)
			suite.Require().Equal(tc.expected, cr.ActionsCompleted)
			suite.Require().True(found)
			suite.Require().NoError(err)

		})
	}
}

func addClaimRecord(ctx sdk.Context, k *claimskeeper.Keeper, actions []bool) sdk.AccAddress {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	cr := claimstypes.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(100), ActionsCompleted: actions}
	k.SetClaimsRecord(ctx, addr, cr)
	return addr
}
