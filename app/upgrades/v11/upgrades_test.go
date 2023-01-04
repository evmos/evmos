package v11_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/evmos/evmos/v10/app"
	v11 "github.com/evmos/evmos/v10/app/upgrades/v11"
	"github.com/evmos/evmos/v10/testutil"
	"github.com/stretchr/testify/suite"

	evmostypes "github.com/evmos/evmos/v10/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
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

func (suite *UpgradeTestSuite) setupEscrowAccounts(accCount int) {
	for i := 0; i <= accCount; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		addr := ibctypes.GetEscrowAddress(ibctypes.PortID, channelID)

		// set accounts as BaseAccounts
		baseAcc := authtypes.NewBaseAccountWithAddress(addr)
		suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)
	}
}

func (suite *UpgradeTestSuite) TestMigrateEscrowAcc() {
	suite.SetupTest(evmostypes.MainnetChainID)

	// fund some escrow accounts
	existingAccounts := 30
	suite.setupEscrowAccounts(existingAccounts)

	// Run migrations
	v11.MigrateEscrowAccounts(suite.ctx, suite.app.AccountKeeper)

	// check account types for channels 0 to 36
	for i := 0; i <= 36; i++ {
		channelID := fmt.Sprintf("channel-%d", i)
		addr := ibctypes.GetEscrowAddress(ibctypes.PortID, channelID)
		acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr)

		if i > existingAccounts {
			suite.Require().Nil(acc, "This account did not exist, it should not be migrated")
			continue
		}
		suite.Require().NotNil(acc)

		moduleAcc, isModuleAccount := acc.(*authtypes.ModuleAccount)
		suite.Require().True(isModuleAccount)
		suite.Require().NoError(moduleAcc.Validate(), "account validation failed")
	}
}

func (suite *UpgradeTestSuite) TestDistributeTestnetRewards() {
	testCases := []struct {
		name            string
		chainID         string
		malleate        func()
		expectedSuccess bool
	}{
		{
			"Mainnet - success",
			evmostypes.MainnetChainID + "-4",
			func() {},
			true,
		},
		{
			"Testnet - no-op",
			evmostypes.TestnetChainID + "-4",
			func() {},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest(evmostypes.MainnetChainID)
			suite.fundTestnetRewardsAcc()

			// call the DistributieTestnetRewards func

			if tc.expectedSuccess {
			} else {
				// addr := sdk.MustAccAddressFromBech32(v11.Accounts[i][0])
				// 	balance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "aevmos")
				// 	suite.Require().Equal(balance.Amount, sdk.NewInt(0))
			}
		})
	}
}

func (suite *UpgradeTestSuite) fundTestnetRewardsAcc() {
	rewardsAcc, err := sdk.AccAddressFromBech32("evmos1f7vxxvmd544dkkmyxan76t76d39k7j3gr8d45y")
	suite.Require().NoError(err)

	amount, ok := sdk.NewIntFromString("7399998994000000000000000")
	suite.Require().True(ok, "error converting rewards account amount")

	rewards := sdk.NewCoins(sdk.NewCoin(evmostypes.BaseDenom, amount))
	err = testutil.FundAccount(suite.ctx, suite.app.BankKeeper, rewardsAcc, rewards)
	suite.Require().NoError(err)
}
