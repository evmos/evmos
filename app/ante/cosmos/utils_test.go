package cosmos_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/evmos/v15/app"
	cosmosante "github.com/evmos/evmos/v15/app/ante/cosmos"
	"github.com/evmos/evmos/v15/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v15/encoding"
	txfactory "github.com/evmos/evmos/v15/testutil/integration/common/factory"
	"github.com/evmos/evmos/v15/utils"
)

func (suite *AnteTestSuite) CreateTestCosmosTxBuilder(gasPrice sdkmath.Int, denom string, msgs ...sdk.Msg) client.TxBuilder {
	txBuilder := suite.network.App.GetTxConfig().NewTxBuilder()

	txBuilder.SetGasLimit(TestGasLimit)
	fees := &sdk.Coins{{Denom: denom, Amount: gasPrice.MulRaw(int64(TestGasLimit))}}
	txBuilder.SetFeeAmount(*fees)
	err := txBuilder.SetMsgs(msgs...)
	suite.Require().NoError(err)
	return txBuilder
}

func (suite *AnteTestSuite) CreateTestCosmosTxBuilderWithFees(fees sdk.Coins, msgs ...sdk.Msg) client.TxBuilder {
	txBuilder := suite.network.App.GetTxConfig().NewTxBuilder()
	txBuilder.SetGasLimit(TestGasLimit)
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
		ChainID:       chainID,
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
func (suite *AnteTestSuite) setupDeductFeeDecoratorTestCase(addr sdk.AccAddress, tc deductFeeDecoratorTestCase) (cosmosante.DeductFeeDecorator, txfactory.CosmosTxArgs) {
	suite.SetupTest()

	// Create a new DeductFeeDecorator
	dfd := cosmosante.NewDeductFeeDecorator(
		suite.network.App.AccountKeeper, suite.network.App.BankKeeper, suite.network.App.DistrKeeper, suite.network.App.FeeGrantKeeper, suite.network.App.StakingKeeper, nil,
	)

	err := suite.prepareAccountsForDelegationRewards(addr, tc.balance, tc.rewards...)
	suite.Require().NoError(err)
	// Create an arbitrary message for testing purposes
	msg := sdktestutil.NewTestMsg(addr)

	// Set up the transaction arguments
	return dfd, txfactory.CosmosTxArgs{
		ChainID:    suite.network.GetChainID(),
		Gas:        &tc.gas,
		GasPrice:   tc.gasPrice,
		FeeGranter: tc.feeGranter,
		Msgs:       []sdk.Msg{msg},
	}
}

// PrepareAccountsForDelegationRewards prepares the test suite for testing to withdraw delegation rewards.
//
// Balance is the amount of tokens that will be left in the account after the setup is done.
// For each defined reward, a validator is created and tokens are allocated to it using the distribution keeper,
// such that the given amount of tokens is outstanding as a staking reward for the account.
//
// The setup is done in the following way:
//   - Fund the account with the given address with the given balance.
//   - If the given balance is zero, the account will be created with zero balance.
//
// For every reward defined in the rewards argument, the following steps are executed:
//   - Set up a validator with zero commission and delegate to it -> the account delegation will be 50% of the total delegation.
//   - Allocate rewards to the validator.
//
// The function returns the updated context along with a potential error.
func (suite *AnteTestSuite) prepareAccountsForDelegationRewards(addr sdk.AccAddress, balance sdkmath.Int, rewards ...sdkmath.Int) error {
	// Calculate the necessary amount of tokens to fund the account in order for the desired residual balance to
	// be left after creating validators and delegating to them.
	totalRewards := math.ZeroInt()
	for _, reward := range rewards {
		totalRewards = totalRewards.Add(reward)
	}
	totalNeededBalance := balance.Add(totalRewards)
	ctx := suite.network.GetContext()
	if totalNeededBalance.IsZero() {
		acc := suite.network.App.AccountKeeper.NewAccountWithAddress(ctx, addr)
		suite.network.App.AccountKeeper.SetAccount(ctx, acc)
	} else {
		// Fund account with enough tokens to stake them
		err := suite.network.FundAccountWithBaseDenom(addr, totalNeededBalance)
		if err != nil {
			return fmt.Errorf("failed to fund account: %s", err.Error())
		}
	}

	if totalRewards.IsZero() {
		return nil
	}

	// set distribution module account balance which pays out the rewards
	err := suite.network.FundModuleAccount(distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, totalRewards)))
	if err != nil {
		return fmt.Errorf("failed to fund distribution module account: %s", err.Error())
	}

	validators := suite.network.GetValidators()
	for i, reward := range rewards {
		if reward.IsZero() {
			continue
		}

		// self-delegate the same amount of tokens as the delegate address also stakes
		// this ensures, that the delegation rewards are 50% of the total rewards
		msgSrv := stakingkeeper.NewMsgServerImpl(&suite.network.App.StakingKeeper)
		coin := sdk.NewCoin(suite.network.GetDenom(), reward)
		msg := stakingtypes.NewMsgDelegate(addr.String(), validators[i].OperatorAddress, coin)
		res, err := msgSrv.Delegate(ctx, msg)
		if err != nil {
			return err
		}

		if res == nil {
			return errors.New("delegation message returned nil response")
		}

		// end block to bond validator and increase block height
		// Not using Commit() here because code panics due to invalid block height
		_, err = suite.network.App.StakingKeeper.EndBlocker(ctx)
		if err != nil {
			return err
		}
		// end block to bond validator and increase block height
		// Not using Commit() here because code panics due to invalid block height
		_, err = suite.network.App.Commit()
		if err != nil {
			return err
		}

		// FIXME getting validator not found when running the anteHandler
		// allocate rewards to validator (of these 50% will be paid out to the delegator)
		allocatedRewards := sdk.NewDecCoins(sdk.NewDecCoin(utils.BaseDenom, reward.Mul(math.NewInt(2))))
		suite.network.App.DistrKeeper.AllocateTokensToValidator(ctx, validators[i], allocatedRewards)
	}

	suite.network.WithContext(ctx)
	return nil
}

// intSlice creates a slice of sdk.Int with the specified size and same value
func intSlice(size int, value sdkmath.Int) []sdkmath.Int {
	slc := make([]sdkmath.Int, size)
	for i := 0; i < len(slc); i++ {
		slc[i] = value
	}
	return slc
}
