package v14rc5_test

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/app/upgrades/v14rc5"
	"github.com/evmos/evmos/v14/crypto/ethsecp256k1"
	distprecompile "github.com/evmos/evmos/v14/precompiles/distribution"
	"github.com/evmos/evmos/v14/utils"
	feemarkettypes "github.com/evmos/evmos/v14/x/feemarket/types"
	"github.com/stretchr/testify/suite"
)

type UpgradeTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.Evmos
	consAddress sdk.ConsAddress
}

func (s *UpgradeTestSuite) SetupTest(chainID string) {
	checkTx := false

	// consensus key
	priv, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err)
	s.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// NOTE: this is the new binary, not the old one.
	s.app = app.Setup(checkTx, feemarkettypes.DefaultGenesisState(), chainID)
	s.ctx = s.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         chainID,
		Time:            time.Date(2022, 5, 9, 8, 0, 0, 0, time.UTC),
		ProposerAddress: s.consAddress.Bytes(),

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

	cp := s.app.BaseApp.GetConsensusParams(s.ctx)
	s.ctx = s.ctx.WithConsensusParams(cp)
}

func TestUpgradeTestSuite(t *testing.T) {
	s := new(UpgradeTestSuite)
	suite.Run(t, s)
}

func (s *UpgradeTestSuite) TestDisableDistributionPrecompile() {
	s.SetupTest(utils.TestnetChainID + "-1")

	initialParams := s.app.EvmKeeper.GetParams(s.ctx)
	distributionPrecompileAddr := distprecompile.Precompile{}.Address().String()
	s.Require().Contains(initialParams.ActivePrecompiles, distributionPrecompileAddr,
		"distribution precompile should be active",
	)

	// run the upgrade logic
	err := v14rc5.DisableDistributionPrecompile(s.ctx, s.app.EvmKeeper)
	s.Require().NoError(err)

	// check that the distribution precompile is no longer active
	updatedParams := s.app.EvmKeeper.GetParams(s.ctx)
	s.Require().NotContains(updatedParams.ActivePrecompiles, distributionPrecompileAddr,
		"distribution precompile should be inactive",
	)
}
