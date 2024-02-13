package cosmos_test

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/evmos/evmos/v16/app"
	cosmosante "github.com/evmos/evmos/v16/app/ante/cosmos"
	"github.com/evmos/evmos/v16/app/ante/testutils"
	"github.com/evmos/evmos/v16/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v16/encoding"
	testutil "github.com/evmos/evmos/v16/testutil"
	"github.com/evmos/evmos/v16/testutil/integration/common/factory"
)

func (suite *AnteTestSuite) CreateTestCosmosTxBuilder(gasPrice sdkmath.Int, denom string, msgs ...sdk.Msg) client.TxBuilder {
	txBuilder := suite.GetClientCtx().TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(testutils.TestGasLimit)
	fees := &sdk.Coins{{Denom: denom, Amount: gasPrice.MulRaw(int64(testutils.TestGasLimit))}}
	txBuilder.SetFeeAmount(*fees)
	err := txBuilder.SetMsgs(msgs...)
	suite.Require().NoError(err)
	return txBuilder
}

func (suite *AnteTestSuite) CreateTestCosmosTxBuilderWithFees(fees sdk.Coins, msgs ...sdk.Msg) client.TxBuilder {
	txBuilder := suite.GetClientCtx().TxConfig.NewTxBuilder()
	txBuilder.SetGasLimit(testutils.TestGasLimit)
	txBuilder.SetFeeAmount(fees)
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

func createTx(ctx context.Context, priv cryptotypes.PrivKey, msgs ...sdk.Msg) (sdk.Tx, error) {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	defaultSignMode, err := authsigning.APISignModeToInternal(encodingConfig.TxConfig.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, err
	}

	txBuilder.SetGasLimit(1000000)
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  defaultSignMode,
			Signature: nil,
		},
		Sequence: 0,
	}

	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		Address:       sdk.AccAddress(priv.PubKey().Bytes()).String(),
		ChainID:       "chainID",
		AccountNumber: 0,
		Sequence:      0,
		PubKey:        priv.PubKey(),
	}

	sigV2, err = tx.SignWithPrivKey(
		ctx, defaultSignMode, signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		0,
	)
	if err != nil {
		return nil, err
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, err
	}

	return txBuilder.GetTx(), nil
}

// setupDeductFeeDecoratorTestCase instantiates a new DeductFeeDecorator
// and prepares the accounts with corresponding balance and staking rewards
// Returns the decorator and the tx arguments to use on the test case
func (suite *AnteTestSuite) setupDeductFeeDecoratorTestCase(addr sdk.AccAddress, priv *ethsecp256k1.PrivKey, tc deductFeeDecoratorTestCase) (sdk.Context, cosmosante.DeductFeeDecorator, factory.CosmosTxArgs) {
	suite.SetupTest()
	nw := suite.GetNetwork()
	ctx := nw.GetContext()

	// Create a new DeductFeeDecorator
	dfd := cosmosante.NewDeductFeeDecorator(
		nw.App.AccountKeeper, nw.App.BankKeeper, nw.App.DistrKeeper, nw.App.FeeGrantKeeper, nw.App.StakingKeeper, nil,
	)

	// prepare the testcase
	var err error
	ctx, err = testutil.PrepareAccountsForDelegationRewards(suite.T(), ctx, nw.App, addr, tc.balance, tc.rewards...)
	suite.Require().NoError(err, "failed to prepare accounts for delegation rewards")

	// Create an arbitrary message for testing purposes
	msg := sdktestutil.NewTestMsg(addr)

	// Set up the transaction arguments
	return ctx, dfd, factory.CosmosTxArgs{
		ChainID:    suite.GetNetwork().GetChainID(),
		Gas:        &tc.gas,
		GasPrice:   tc.gasPrice,
		FeeGranter: tc.feeGranter,
		Msgs:       []sdk.Msg{msg},
	}
}

// intSlice creates a slice of sdkmath.Int with the specified size and same value
func intSlice(size int, value sdkmath.Int) []sdkmath.Int {
	slc := make([]sdkmath.Int, size)
	for i := 0; i < len(slc); i++ {
		slc[i] = value
	}
	return slc
}
