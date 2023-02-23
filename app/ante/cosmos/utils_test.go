package cosmos_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/encoding"
	"github.com/evmos/evmos/v11/testutil"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/utils"
)

func (suite *AnteTestSuite) CreateTestCosmosTxBuilder(gasPrice sdkmath.Int, denom string, msgs ...sdk.Msg) client.TxBuilder {
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(TestGasLimit)
	fees := &sdk.Coins{{Denom: denom, Amount: gasPrice.MulRaw(int64(TestGasLimit))}}
	txBuilder.SetFeeAmount(*fees)
	err := txBuilder.SetMsgs(msgs...)
	suite.Require().NoError(err)
	return txBuilder
}

func newMsgExec(grantee sdk.AccAddress, msgs []sdk.Msg) *authz.MsgExec {
	msg := authz.NewMsgExec(grantee, msgs)
	return &msg
}

func newMsgGrant(granter sdk.AccAddress, grantee sdk.AccAddress, a authz.Authorization, expiration *time.Time) *authz.MsgGrant {
	msg, err := authz.NewMsgGrant(granter, grantee, a, expiration)
	if err != nil {
		panic(err)
	}
	return msg
}

func createNestedMsgExec(a sdk.AccAddress, nestedLvl int, lastLvlMsgs []sdk.Msg) *authz.MsgExec {
	msgs := make([]*authz.MsgExec, nestedLvl)
	for i := range msgs {
		if i == 0 {
			msgs[i] = newMsgExec(a, lastLvlMsgs)
			continue
		}
		msgs[i] = newMsgExec(a, []sdk.Msg{msgs[i-1]})
	}
	return msgs[nestedLvl-1]
}

func generatePrivKeyAddressPairs(accCount int) ([]*ethsecp256k1.PrivKey, []sdk.AccAddress, error) {
	var (
		err           error
		testPrivKeys  = make([]*ethsecp256k1.PrivKey, accCount)
		testAddresses = make([]sdk.AccAddress, accCount)
	)

	for i := range testPrivKeys {
		testPrivKeys[i], err = ethsecp256k1.GenerateKey()
		if err != nil {
			return nil, nil, err
		}
		testAddresses[i] = testPrivKeys[i].PubKey().Address().Bytes()
	}
	return testPrivKeys, testAddresses, nil
}

func createTx(priv cryptotypes.PrivKey, msgs ...sdk.Msg) (sdk.Tx, error) {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(1000000)
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: 0,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: 0,
		Sequence:      0,
	}
	sigV2, err := tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		0,
	)
	if err != nil {
		return nil, err
	}

	sigsV2 = []signing.SignatureV2{sigV2}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return txBuilder.GetTx(), nil
}

// PrepareAccountsForDelegationRewards prepares the test suite for testing to withdraw delegation rewards.
//
// The setup is done in the following way:
//   - Fund the account with the given address with the given balance.
//     If the given balance is zero, the account will be created with zero balance.
//   - Set up a validator with zero commission and delegate to it -> the account delegation will be 50% of the total delegation.
//   - Allocate rewards to the validator.
func PrepareAccountsForDelegationRewards(suite *AnteTestSuite, addr sdk.AccAddress, balance, rewards sdkmath.Int) {
	// reset historical count in distribution keeper which is necessary
	// for the delegation rewards to be calculated correctly
	suite.app.DistrKeeper.DeleteAllValidatorHistoricalRewards(suite.ctx)

	if balance.IsZero() {
		suite.app.AccountKeeper.SetAccount(suite.ctx, suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
	} else {
		// Fund account with enough tokens to stake them
		err := testutil.FundAccountWithBaseDenom(suite.ctx, suite.app.BankKeeper, addr, balance.Int64())
		suite.Require().NoError(err, "failed to fund account")
	}

	if !rewards.IsZero() {
		// Set up validator and delegate to it
		privKey := ed25519.GenPrivKey()
		addr2, _ := testutiltx.NewAccAddressAndKey()
		err := testutil.FundAccountWithBaseDenom(suite.ctx, suite.app.BankKeeper, addr2, rewards.Int64())
		suite.Require().NoError(err, "failed to fund validator account")

		zeroDec := sdk.ZeroDec()
		stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
		stakingParams.BondDenom = utils.BaseDenom
		stakingParams.MinCommissionRate = zeroDec
		suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

		stakingHelper := teststaking.NewHelper(suite.T(), suite.ctx, suite.app.StakingKeeper)
		stakingHelper.Commission = stakingtypes.NewCommissionRates(zeroDec, zeroDec, zeroDec)
		stakingHelper.Denom = utils.BaseDenom

		valAddr := sdk.ValAddress(addr2.Bytes())
		// self-delegate the same amount of tokens as the delegate address also stakes
		// this ensures, that the delegation rewards are 50% of the total rewards
		stakingHelper.CreateValidator(valAddr, privKey.PubKey(), rewards, true)
		stakingHelper.Delegate(addr, valAddr, rewards)

		// TODO: Replace this with testutil.Commit?
		// end block to bond validator and increase block height
		staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

		// set distribution module account balance which pays out the rewards
		distrAcc := suite.app.DistrKeeper.GetDistributionAccount(suite.ctx)
		err = testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, rewards)))
		suite.Require().NoError(err, "failed to fund distribution module account")
		suite.app.AccountKeeper.SetModuleAccount(suite.ctx, distrAcc)

		// allocate rewards to validator (of these 50% will be paid out to the delegator)
		validator := suite.app.StakingKeeper.Validator(suite.ctx, valAddr)
		allocatedRewards := sdk.NewDecCoins(sdk.NewDecCoin(utils.BaseDenom, rewards.Mul(sdk.NewInt(2))))
		suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, allocatedRewards)
	}
}
