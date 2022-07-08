package feesplit_test

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

	"github.com/evmos/ethermint/tests"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/evmos/evmos/v6/app"
	"github.com/evmos/evmos/v6/x/feesplit"
	"github.com/evmos/evmos/v6/x/feesplit/types"
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

	suite.genesis = *types.DefaultGenesisState()
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestFeesplitInitGenesis() {
	testCases := []struct {
		name     string
		genesis  types.GenesisState
		expPanic bool
	}{
		{
			"default genesis",
			suite.genesis,
			false,
		},
		{
			"custom genesis - feesplit disabled",
			types.GenesisState{
				Params: types.Params{
					EnableFeeSplit:           false,
					DeveloperShares:          types.DefaultDeveloperShares,
					AddrDerivationCostCreate: types.DefaultAddrDerivationCostCreate,
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
					feesplit.InitGenesis(suite.ctx, suite.app.FeesplitKeeper, tc.genesis)
				})
			} else {
				suite.Require().NotPanics(func() {
					feesplit.InitGenesis(suite.ctx, suite.app.FeesplitKeeper, tc.genesis)
				})

				params := suite.app.FeesplitKeeper.GetParams(suite.ctx)
				suite.Require().Equal(tc.genesis.Params, params)
			}
		})
	}
}

func (suite *GenesisTestSuite) TestFeesplitExportGenesis() {
	feesplit.InitGenesis(suite.ctx, suite.app.FeesplitKeeper, suite.genesis)

	genesisExported := feesplit.ExportGenesis(suite.ctx, suite.app.FeesplitKeeper)
	suite.Require().Equal(genesisExported.Params, suite.genesis.Params)
}
