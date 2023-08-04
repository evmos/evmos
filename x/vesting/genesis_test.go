package vesting_test

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

	"github.com/evmos/evmos/v13/app"
	utiltx "github.com/evmos/evmos/v13/testutil/tx"
	"github.com/evmos/evmos/v13/utils"
	feemarkettypes "github.com/evmos/evmos/v13/x/feemarket/types"
	"github.com/evmos/evmos/v13/x/vesting"
	"github.com/evmos/evmos/v13/x/vesting/types"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app     *app.Evmos
	genesis types.GenesisState
}

func (suite *GenesisTestSuite) SetupTest() {
	// consensus key
	consAddress := sdk.ConsAddress(utiltx.GenerateAddress().Bytes())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         utils.TestnetChainID + "-1",
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

	suite.genesis = *types.DefaultGenesisState()
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestVestingInitGenesis() {
	testCases := []struct {
		name     string
		genesis  types.GenesisState
		expPanic bool
	}{
		{
			"pass - default genesis",
			suite.genesis,
			false,
		},
		{
			"pass - custom genesis - gov clawback disabled",
			types.GenesisState{
				Params: types.Params{
					EnableGovClawback: false,
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			if tc.expPanic {
				suite.Require().Panics(func() {
					vesting.InitGenesis(suite.ctx, suite.app.VestingKeeper, tc.genesis)
				})
			} else {
				suite.Require().NotPanics(func() {
					vesting.InitGenesis(suite.ctx, suite.app.VestingKeeper, tc.genesis)
				})

				params := suite.app.VestingKeeper.GetParams(suite.ctx)
				suite.Require().Equal(tc.genesis.Params, params)
			}
		})
	}
}

func (suite *GenesisTestSuite) TestVestingExportGenesis() {
	vesting.InitGenesis(suite.ctx, suite.app.VestingKeeper, suite.genesis)

	genesisExported := vesting.ExportGenesis(suite.ctx, suite.app.VestingKeeper)
	suite.Require().Equal(genesisExported.Params, suite.genesis.Params)
}
