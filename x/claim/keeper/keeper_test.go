package keeper_test

import (
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
	cmdcfg "github.com/tharsis/evmos/cmd/config"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	"github.com/tharsis/ethermint/crypto/ethsecp256k1"

	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/claim/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app         *app.Evmos
	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	consAddress := sdk.ConsAddress(priv.PubKey().Address())

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

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.ClaimKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	params := types.DefaultParams()
	params.AirdropStartTime = suite.ctx.BlockTime()
	suite.app.ClaimKeeper.SetParams(suite.ctx, params)

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = params.GetClaimDenom()
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	govParams := suite.app.GovKeeper.GetDepositParams(suite.ctx)
	govParams.MinDeposit[0].Denom = params.GetClaimDenom()
}

func init() {
	config := sdk.GetConfig()
	cmdcfg.SetBech32Prefixes(config)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
