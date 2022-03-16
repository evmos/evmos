package claims_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	"github.com/tharsis/ethermint/tests"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/tharsis/evmos/v2/app"
	"github.com/tharsis/evmos/v2/x/claims"
	"github.com/tharsis/evmos/v2/x/claims/types"
	inflationtypes "github.com/tharsis/evmos/v2/x/inflation/types"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app     *app.Evmos
	genesis types.GenesisState
}

func (suite *GenesisTestSuite) SetupTest() {
	// consensus key
	consAddress := sdk.ConsAddress(tests.GenerateAddress().Bytes())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9000-1",
		Time:            time.Now().UTC(),
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

	params := types.DefaultParams()
	params.AirdropStartTime = suite.ctx.BlockTime()
	suite.app.ClaimsKeeper.SetParams(suite.ctx, params)

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = params.GetClaimsDenom()
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	suite.genesis = *types.DefaultGenesis()
	suite.genesis.Params.AirdropStartTime = suite.ctx.BlockTime()
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

var (
	now     = time.Now().UTC()
	acc1, _ = sdk.AccAddressFromBech32("evmos1qxx0fdsmruzuar2fay88lfw6sce6emamyu2s8h4d")
	acc2, _ = sdk.AccAddressFromBech32("evmos1nsrs4t7dngkdltehkm3p6n8dp22sz3mct9uhc8")
)

func (suite *GenesisTestSuite) TestClaimInitGenesis() {
	testCases := []struct {
		name     string
		genesis  types.GenesisState
		malleate func()
		expPanic bool
	}{
		{
			"default genesis",
			suite.genesis,
			func() {},
			false,
		},
		{
			"custom genesis - not all claimed",
			types.GenesisState{
				Params: suite.genesis.Params,
				ClaimsRecords: []types.ClaimsRecordAddress{
					{
						Address:                acc1.String(),
						InitialClaimableAmount: sdk.NewInt(10_000),
						ActionsCompleted:       []bool{true, false, true, true},
					},
					{
						Address:                acc2.String(),
						InitialClaimableAmount: sdk.NewInt(400),
						ActionsCompleted:       []bool{false, false, true, false},
					},
				},
			},
			func() {
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(2_800)))
				err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, inflationtypes.ModuleName, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"custom genesis - all claimed or all unclaimed",
			types.GenesisState{
				Params: suite.genesis.Params,
				ClaimsRecords: []types.ClaimsRecordAddress{
					{
						Address:                acc1.String(),
						InitialClaimableAmount: sdk.NewInt(10_000),
						ActionsCompleted:       []bool{true, true, true, true},
					},
					{
						Address:                acc2.String(),
						InitialClaimableAmount: sdk.NewInt(400),
						ActionsCompleted:       []bool{false, false, false, false},
					},
				},
			},
			func() {
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(400)))
				err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, inflationtypes.ModuleName, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			if tc.expPanic {
				suite.Require().Panics(func() {
					claims.InitGenesis(suite.ctx, suite.app.ClaimsKeeper, tc.genesis)
				})
			} else {
				suite.Require().NotPanics(func() {
					claims.InitGenesis(suite.ctx, suite.app.ClaimsKeeper, tc.genesis)
				})

				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				suite.Require().Equal(params, tc.genesis.Params)

				claimsRecords := suite.app.ClaimsKeeper.GetClaimsRecords(suite.ctx)
				suite.Require().Equal(claimsRecords, tc.genesis.ClaimsRecords)
			}
		})
	}
}

func (suite *GenesisTestSuite) TestClaimExportGenesis() {
	suite.genesis.ClaimsRecords = []types.ClaimsRecordAddress{
		{
			Address:                acc1.String(),
			InitialClaimableAmount: sdk.NewInt(10_000),
			ActionsCompleted:       []bool{true, true, true, true},
		},
		{
			Address:                acc2.String(),
			InitialClaimableAmount: sdk.NewInt(400),
			ActionsCompleted:       []bool{false, false, false, false},
		},
	}

	coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(400)))
	err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, inflationtypes.ModuleName, types.ModuleName, coins)
	suite.Require().NoError(err)

	claims.InitGenesis(suite.ctx, suite.app.ClaimsKeeper, suite.genesis)

	claimsRecord, found := suite.app.ClaimsKeeper.GetClaimsRecord(suite.ctx, acc2)
	suite.Require().True(found)
	suite.Require().Equal(claimsRecord, types.ClaimsRecord{
		InitialClaimableAmount: sdk.NewInt(400),
		ActionsCompleted:       []bool{false, false, false, false},
	})

	claimableAmount := suite.app.ClaimsKeeper.GetClaimableAmountForAction(suite.ctx, claimsRecord, types.ActionIBCTransfer, suite.genesis.Params)
	suite.Require().Equal(claimableAmount, sdk.NewInt(100))

	genesisExported := claims.ExportGenesis(suite.ctx, suite.app.ClaimsKeeper)
	suite.Require().Equal(genesisExported.Params, suite.genesis.Params)
	suite.Require().Equal(genesisExported.ClaimsRecords, suite.genesis.ClaimsRecords)
}
