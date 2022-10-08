package v83_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/evmos/evmos/v9/app"
	v83 "github.com/evmos/evmos/v9/app/upgrades/v8_3"
	evmostypes "github.com/evmos/evmos/v9/types"
	"github.com/evmos/evmos/v9/x/erc20/types"
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

func (suite *UpgradeTestSuite) TestMigrateIBCModuleAccount() {

	suite.SetupTest(evmostypes.TestnetChainID + "-2") // reset

	// SEND FUNDS TO THE COMMUNITY POOL
	priv, err := ethsecp256k1.GenerateKey()
	address := common.BytesToAddress(priv.PubKey().Address().Bytes())
	sender := sdk.AccAddress(address.Bytes())
	res, _ := sdkmath.NewIntFromString("73575669925896300000000")
	coins := sdk.NewCoins(sdk.NewCoin("aevmos", res))
	suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
	suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
	err = suite.app.DistrKeeper.FundCommunityPool(suite.ctx, coins, sender)
	suite.Require().NoError(err)

	// RETURN FUNDS TO ACCOUNTS AFFECTED
	v83.ReturnFundsFromCommunityPool(suite.ctx, suite.app.DistrKeeper)

	// CHECK BALANCE OF AFFECTED ACCOUNTS
	for account, amount := range v83.Accounts {
		addr := sdk.MustAccAddressFromBech32(account)
		res, _ := sdkmath.NewIntFromString(amount)
		balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "aevmos")
		suite.Require().Equal(balance.Amount, res)
	}
}
