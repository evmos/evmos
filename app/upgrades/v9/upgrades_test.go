package v9_test

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/evmos/v11/x/feemarket/types"

	"github.com/evmos/evmos/v11/app"
	v9 "github.com/evmos/evmos/v11/app/upgrades/v9"
	"github.com/evmos/evmos/v11/utils"
	"github.com/evmos/evmos/v11/x/erc20/types"
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

func (suite *UpgradeTestSuite) TestReturnFundsFromCommunityPool() {
	suite.SetupTest(utils.TestnetChainID + "-2")

	// send funds to the community pool
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	address := common.BytesToAddress(priv.PubKey().Address().Bytes())
	sender := sdk.AccAddress(address.Bytes())
	res, _ := sdk.NewIntFromString(v9.MaxRecover)
	coins := sdk.NewCoins(sdk.NewCoin("aevmos", res))
	err = suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
	suite.Require().NoError(err)
	err = suite.app.DistrKeeper.FundCommunityPool(suite.ctx, coins, sender)
	suite.Require().NoError(err)

	balanceBefore := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Require().Equal(balanceBefore.AmountOf("aevmos"), sdk.NewDecFromInt(res))

	// return funds to accounts affected
	err = v9.ReturnFundsFromCommunityPool(suite.ctx, suite.app.DistrKeeper)
	suite.Require().NoError(err)

	// check balance of affected accounts
	for i := range v9.Accounts {
		addr := sdk.MustAccAddressFromBech32(v9.Accounts[i][0])
		res, _ := sdk.NewIntFromString(v9.Accounts[i][1])
		balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "aevmos")
		suite.Require().Equal(balance.Amount, res)
	}

	balanceAfter := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)
	suite.Require().True(balanceAfter.IsZero())
}
