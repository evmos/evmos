package v14rc2_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v13/app"
	"github.com/evmos/evmos/v13/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v13/testutil"
	"github.com/evmos/evmos/v13/utils"
	feemarkettypes "github.com/evmos/evmos/v13/x/feemarket/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"testing"
	"time"
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
	s.app = app.Setup(checkTx, feemarkettypes.DefaultGenesisState())
	s.ctx = s.app.BaseApp.NewContext(
		checkTx,
		testutil.NewHeader(
			1,
			time.Now(),
			chainID,
			s.consAddress.Bytes(),
			tmhash.Sum([]byte("block_id")),
			tmhash.Sum([]byte("validators")),
		),
	)

	cp := s.app.BaseApp.GetConsensusParams(s.ctx)
	s.ctx = s.ctx.WithConsensusParams(cp)
}

func TestUpgradeTestSuite(t *testing.T) {
	s := new(UpgradeTestSuite)
	suite.Run(t, s)
}

func (s *UpgradeTestSuite) TestUpdateVestingFunders() {
	s.SetupTest(utils.TestnetChainID + "-2")
	s.Require().True(false)
}
