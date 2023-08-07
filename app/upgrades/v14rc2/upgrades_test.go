package v14rc2_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v13/app"
	"github.com/evmos/evmos/v13/app/upgrades/v14rc2"
	"github.com/evmos/evmos/v13/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v13/testutil"
	"github.com/evmos/evmos/v13/utils"
	feemarkettypes "github.com/evmos/evmos/v13/x/feemarket/types"
	"github.com/evmos/evmos/v13/x/vesting/types"
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
}

func TestUpgradeTestSuite(t *testing.T) {
	s := new(UpgradeTestSuite)
	suite.Run(t, s)
}

func (s *UpgradeTestSuite) TestUpdateVestingFunders() {
	s.SetupTest(utils.TestnetChainID + "-2")

	// Fund the affected accounts to initialize them and then create vesting accounts
	for address, oldFunder := range v14rc2.AffectedAddresses {
		accAddr := sdk.MustAccAddressFromBech32(address)
		err := testutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, accAddr, 1000)
		s.Require().NoError(err, "failed to fund account %s", address)

		// Create vesting account
		createMsg := &types.MsgCreateClawbackVestingAccount{
			FunderAddress:  oldFunder,
			VestingAddress: address,
		}
		_, err = s.app.VestingKeeper.CreateClawbackVestingAccount(sdk.UnwrapSDKContext(s.ctx), createMsg)
		s.Require().NoError(err, "failed to create vesting account for %s", address)
	}

	// Run the upgrade function
	err := v14rc2.UpdateVestingFunders(s.ctx, s.app.VestingKeeper)
	s.Require().NoError(err, "failed to update vesting funders")

	// Check that the vesting accounts have been updated
	for address := range v14rc2.AffectedAddresses {
		accAddr := sdk.MustAccAddressFromBech32(address)
		acc := s.app.AccountKeeper.GetAccount(s.ctx, accAddr)
		s.Require().NotNil(acc, "account not found for %s", address)
		vestingAcc, ok := acc.(*types.ClawbackVestingAccount)
		s.Require().True(ok, "account is not a vesting account for %s", address)
		s.Require().Equal(address, vestingAcc.Address, "expected different address in vesting account for %s", address)

		// Check that the funder has been updated
		s.Require().Equal(v14rc2.NewTeamMultisigAcc.String(), vestingAcc.FunderAddress, "expected different funder address for %s", address)
	}
}
