package evm_test

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evtypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/ethereum/eip712"
	"github.com/evmos/evmos/v11/testutil"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/utils"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"
)

func (suite *AnteTestSuite) BuildTestEthTx(
	from common.Address,
	to common.Address,
	amount *big.Int,
	input []byte,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	accesses *ethtypes.AccessList,
) *evmtypes.MsgEthereumTx {
	chainID := suite.app.EvmKeeper.ChainID()
	nonce := suite.app.EvmKeeper.GetNonce(
		suite.ctx,
		common.BytesToAddress(from.Bytes()),
	)

	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &to,
		Amount:    amount,
		GasLimit:  TestGasLimit,
		GasPrice:  gasPrice,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Input:     input,
		Accesses:  accesses,
	}

	msgEthereumTx := evmtypes.NewTx(ethTxParams)
	msgEthereumTx.From = from.String()
	return msgEthereumTx
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (suite *AnteTestSuite) CreateTestTx(
	msg *evmtypes.MsgEthereumTx, priv cryptotypes.PrivKey, accNum uint64, signCosmosTx bool,
	unsetExtensionOptions ...bool,
) authsigning.Tx {
	return suite.CreateTestTxBuilder(msg, priv, accNum, signCosmosTx).GetTx()
}

// CreateTestTxBuilder is a helper function to create a tx builder given multiple inputs.
func (suite *AnteTestSuite) CreateTestTxBuilder(
	msg *evmtypes.MsgEthereumTx, priv cryptotypes.PrivKey, accNum uint64, signCosmosTx bool,
	unsetExtensionOptions ...bool,
) client.TxBuilder {
	var option *codectypes.Any
	var err error
	if len(unsetExtensionOptions) == 0 {
		option, err = codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
		suite.Require().NoError(err)
	}

	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok)

	if len(unsetExtensionOptions) == 0 {
		builder.SetExtensionOptions(option)
	}

	err = msg.Sign(suite.ethSigner, testutiltx.NewSigner(priv))
	suite.Require().NoError(err)

	msg.From = ""
	err = builder.SetMsgs(msg)
	suite.Require().NoError(err)

	txData, err := evmtypes.UnpackTxData(msg.Data)
	suite.Require().NoError(err)

	fees := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewIntFromBigInt(txData.Fee())))
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msg.GetGas())

	if signCosmosTx {
		// First round: we gather all the signer infos. We use the "set empty
		// signature" hack to do that.
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: txData.GetNonce(),
		}

		sigsV2 := []signing.SignatureV2{sigV2}

		err = txBuilder.SetSignatures(sigsV2...)
		suite.Require().NoError(err)

		// Second round: all signer infos are set, so each signer can sign.

		signerData := authsigning.SignerData{
			ChainID:       suite.ctx.ChainID(),
			AccountNumber: accNum,
			Sequence:      txData.GetNonce(),
		}
		sigV2, err = tx.SignWithPrivKey(
			suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			txBuilder, priv, suite.clientCtx.TxConfig, txData.GetNonce(),
		)
		suite.Require().NoError(err)

		sigsV2 = []signing.SignatureV2{sigV2}

		err = txBuilder.SetSignatures(sigsV2...)
		suite.Require().NoError(err)
	}

	return txBuilder
}

func (suite *AnteTestSuite) CreateTestCosmosTxBuilder(gasPrice sdkmath.Int, denom string, msgs ...sdk.Msg) client.TxBuilder {
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(TestGasLimit)
	fees := &sdk.Coins{{Denom: denom, Amount: gasPrice.MulRaw(int64(TestGasLimit))}}
	txBuilder.SetFeeAmount(*fees)
	err := txBuilder.SetMsgs(msgs...)
	suite.Require().NoError(err)
	return txBuilder
}

func (suite *AnteTestSuite) CreateTestEIP712TxBuilderMsgSend(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	// Build MsgSend
	recipient := sdk.AccAddress(common.Address{}.Bytes())
	msgSend := banktypes.NewMsgSend(from, recipient, sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(1))))
	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, msgSend)
}

func (suite *AnteTestSuite) CreateTestEIP712TxBuilderMsgDelegate(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	// Build MsgSend
	valEthAddr := testutiltx.GenerateAddress()
	valAddr := sdk.ValAddress(valEthAddr.Bytes())
	msgSend := stakingtypes.NewMsgDelegate(from, valAddr, sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20)))
	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, msgSend)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgCreateValidator(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	// Build MsgCreateValidator
	valAddr := sdk.ValAddress(from.Bytes())
	privEd := ed25519.GenPrivKey()
	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		valAddr,
		privEd.PubKey(),
		sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(20)),
		stakingtypes.NewDescription("moniker", "indentity", "website", "security_contract", "details"),
		stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
		sdk.OneInt(),
	)
	suite.Require().NoError(err)
	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, msgCreate)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgCreateValidator2(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	// Build MsgCreateValidator
	valAddr := sdk.ValAddress(from.Bytes())
	privEd := ed25519.GenPrivKey()
	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		valAddr,
		privEd.PubKey(),
		sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(20)),
		// Ensure optional fields can be left blank
		stakingtypes.NewDescription("moniker", "indentity", "", "", ""),
		stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
		sdk.OneInt(),
	)
	suite.Require().NoError(err)
	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, msgCreate)
}

func (suite *AnteTestSuite) CreateTestEIP712SubmitProposal(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins, deposit sdk.Coins) (client.TxBuilder, error) {
	proposal, ok := govtypes.ContentFromProposalType("My proposal", "My description", govtypes.ProposalTypeText)
	suite.Require().True(ok)
	msgSubmit, err := govtypes.NewMsgSubmitProposal(proposal, deposit, from)
	suite.Require().NoError(err)
	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, msgSubmit)
}

func (suite *AnteTestSuite) CreateTestEIP712GrantAllowance(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	spendLimit := sdk.NewCoins(sdk.NewInt64Coin(evmtypes.DefaultEVMDenom, 10))
	threeHours := time.Now().Add(3 * time.Hour)
	basic := &feegrant.BasicAllowance{
		SpendLimit: spendLimit,
		Expiration: &threeHours,
	}
	granted := testutiltx.GenerateAddress()
	grantedAddr := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, granted.Bytes())
	msgGrant, err := feegrant.NewMsgGrantAllowance(basic, from, grantedAddr.GetAddress())
	suite.Require().NoError(err)
	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, msgGrant)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgEditValidator(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	valAddr := sdk.ValAddress(from.Bytes())
	msgEdit := stakingtypes.NewMsgEditValidator(
		valAddr,
		stakingtypes.NewDescription("moniker", "identity", "website", "security_contract", "details"),
		nil,
		nil,
	)
	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, msgEdit)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgSubmitEvidence(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	pk := ed25519.GenPrivKey()
	msgEvidence, err := evtypes.NewMsgSubmitEvidence(from, &evtypes.Equivocation{
		Height:           11,
		Time:             time.Now().UTC(),
		Power:            100,
		ConsensusAddress: pk.PubKey().Address().String(),
	})
	suite.Require().NoError(err)

	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, msgEvidence)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgVoteV1(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	msgVote := govtypesv1.NewMsgVote(from, 1, govtypesv1.VoteOption_VOTE_OPTION_YES, "")
	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, msgVote)
}

func (suite *AnteTestSuite) CreateTestEIP712SubmitProposalV1(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	// Build V1 proposal messages. Must all be same-type, since EIP-712
	// does not support arrays of variable type.
	authAcc := suite.app.GovKeeper.GetGovernanceAccount(suite.ctx)

	proposal1, ok := govtypes.ContentFromProposalType("My proposal 1", "My description 1", govtypes.ProposalTypeText)
	suite.Require().True(ok)
	content1, err := govtypesv1.NewLegacyContent(
		proposal1,
		sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), authAcc.GetAddress().Bytes()),
	)
	suite.Require().NoError(err)

	proposal2, ok := govtypes.ContentFromProposalType("My proposal 2", "My description 2", govtypes.ProposalTypeText)
	suite.Require().True(ok)
	content2, err := govtypesv1.NewLegacyContent(
		proposal2,
		sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), authAcc.GetAddress().Bytes()),
	)
	suite.Require().NoError(err)

	proposalMsgs := []sdk.Msg{
		content1,
		content2,
	}

	// Build V1 proposal
	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		proposalMsgs,
		sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100))),
		sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), from.Bytes()),
		"Metadata",
	)

	suite.Require().NoError(err)

	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, msgProposal)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgExec(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	recipient := sdk.AccAddress(common.Address{}.Bytes())
	msgSend := banktypes.NewMsgSend(from, recipient, sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(1))))
	msgExec := authz.NewMsgExec(from, []sdk.Msg{msgSend})
	return suite.CreateTestEIP712SingleMessageTxBuilder(from, priv, chainID, gas, gasAmount, &msgExec)
}

func (suite *AnteTestSuite) CreateTestEIP712MultipleMsgSend(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	recipient := sdk.AccAddress(common.Address{}.Bytes())
	msgSend := banktypes.NewMsgSend(from, recipient, sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(1))))
	txConf := suite.clientCtx.TxConfig
	msgs := []sdk.Msg{msgSend, msgSend, msgSend}
	return testutiltx.PrepareEIP712CosmosTx(
		suite.ctx,
		suite.app,
		testutiltx.CosmosTxArgs{
			TxCfg:   txConf,
			Priv:    priv,
			ChainID: chainID,
			Gas:     gas,
			Fees:    gasAmount,
			Msgs:    msgs,
		},
	)
}

// Fails
func (suite *AnteTestSuite) CreateTestEIP712MultipleSignerMsgs(from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins) (client.TxBuilder, error) {
	recipient := sdk.AccAddress(common.Address{}.Bytes())
	msgSend1 := banktypes.NewMsgSend(from, recipient, sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(1))))
	msgSend2 := banktypes.NewMsgSend(recipient, from, sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(1))))
	txConf := suite.clientCtx.TxConfig
	msgs := []sdk.Msg{msgSend1, msgSend2}
	return testutiltx.PrepareEIP712CosmosTx(
		suite.ctx,
		suite.app,
		testutiltx.CosmosTxArgs{
			TxCfg:   txConf,
			Priv:    priv,
			ChainID: chainID,
			Gas:     gas,
			Fees:    gasAmount,
			Msgs:    msgs,
		},
	)
}

// StdSignBytes returns the bytes to sign for a transaction.
func StdSignBytes(cdc *codec.LegacyAmino, chainID string, accnum uint64, sequence uint64, timeout uint64, fee legacytx.StdFee, msgs []sdk.Msg, memo string, tip *txtypes.Tip) []byte {
	msgsBytes := make([]json.RawMessage, 0, len(msgs))
	for _, msg := range msgs {
		legacyMsg, ok := msg.(legacytx.LegacyMsg)
		if !ok {
			panic(fmt.Errorf("expected %T when using amino JSON", (*legacytx.LegacyMsg)(nil)))
		}

		msgsBytes = append(msgsBytes, json.RawMessage(legacyMsg.GetSignBytes()))
	}

	var stdTip *legacytx.StdTip
	if tip != nil {
		if tip.Tipper == "" {
			panic(fmt.Errorf("tipper cannot be empty"))
		}

		stdTip = &legacytx.StdTip{Amount: tip.Amount, Tipper: tip.Tipper}
	}

	bz, err := cdc.MarshalJSON(legacytx.StdSignDoc{
		AccountNumber: accnum,
		ChainID:       chainID,
		Fee:           json.RawMessage(fee.Bytes()),
		Memo:          memo,
		Msgs:          msgsBytes,
		Sequence:      sequence,
		TimeoutHeight: timeout,
		Tip:           stdTip,
	})
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bz)
}

func (suite *AnteTestSuite) CreateTestEIP712SingleMessageTxBuilder(
	from sdk.AccAddress, priv cryptotypes.PrivKey, chainID string, gas uint64, gasAmount sdk.Coins, msg sdk.Msg,
) (client.TxBuilder, error) {
	txConf := suite.clientCtx.TxConfig
	msgs := []sdk.Msg{msg}
	return testutiltx.PrepareEIP712CosmosTx(
		suite.ctx,
		suite.app,
		testutiltx.CosmosTxArgs{
			TxCfg:   txConf,
			Priv:    priv,
			ChainID: chainID,
			Gas:     gas,
			Fees:    gasAmount,
			Msgs:    msgs,
		},
	)
}

// Generate a set of pub/priv keys to be used in creating multi-keys
func (suite *AnteTestSuite) GenerateMultipleKeys(n int) ([]cryptotypes.PrivKey, []cryptotypes.PubKey) {
	privKeys := make([]cryptotypes.PrivKey, n)
	pubKeys := make([]cryptotypes.PubKey, n)
	for i := 0; i < n; i++ {
		privKey, err := ethsecp256k1.GenerateKey()
		suite.Require().NoError(err)
		privKeys[i] = privKey
		pubKeys[i] = privKey.PubKey()
	}
	return privKeys, pubKeys
}

// generateSingleSignature signs the given sign doc bytes using the given signType (EIP-712 or Standard)
func (suite *AnteTestSuite) generateSingleSignature(signMode signing.SignMode, privKey cryptotypes.PrivKey, signDocBytes []byte, signType string) (signature signing.SignatureV2) {
	var (
		msg []byte
		err error
	)

	msg = signDocBytes

	if signType == "EIP-712" {
		msg, err = eip712.GetEIP712BytesForMsg(signDocBytes)
		suite.Require().NoError(err)
	}

	sigBytes, _ := privKey.Sign(msg)
	sigData := &signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: sigBytes,
	}

	return signing.SignatureV2{
		PubKey: privKey.PubKey(),
		Data:   sigData,
	}
}

// generateMultikeySignatures signs a set of messages using each private key within a given multi-key
func (suite *AnteTestSuite) generateMultikeySignatures(signMode signing.SignMode, privKeys []cryptotypes.PrivKey, signDocBytes []byte, signType string) (signatures []signing.SignatureV2) {
	n := len(privKeys)
	signatures = make([]signing.SignatureV2, n)

	for i := 0; i < n; i++ {
		privKey := privKeys[i]
		currentType := signType

		// If mixed type, alternate signing type on each iteration
		if signType == "mixed" {
			if i%2 == 0 {
				currentType = "EIP-712"
			} else {
				currentType = "Standard"
			}
		}

		signatures[i] = suite.generateSingleSignature(
			signMode,
			privKey,
			signDocBytes,
			currentType,
		)
	}

	return signatures
}

// RegisterAccount creates an account with the keeper and populates the initial balance
func (suite *AnteTestSuite) RegisterAccount(pubKey cryptotypes.PubKey, balance *big.Int) {
	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, sdk.AccAddress(pubKey.Address()))
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	err := suite.app.EvmKeeper.SetBalance(suite.ctx, common.BytesToAddress(pubKey.Address()), balance)
	suite.Require().NoError(err)
}

// createSignerBytes generates sign doc bytes using the given parameters
func (suite *AnteTestSuite) createSignerBytes(chainID string, signMode signing.SignMode, pubKey cryptotypes.PubKey, txBuilder client.TxBuilder) []byte {
	acc, err := sdkante.GetSignerAcc(suite.ctx, suite.app.AccountKeeper, sdk.AccAddress(pubKey.Address()))
	suite.Require().NoError(err)
	signerInfo := authsigning.SignerData{
		Address:       sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), acc.GetAddress().Bytes()),
		ChainID:       chainID,
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
		PubKey:        pubKey,
	}

	signerBytes, err := suite.clientCtx.TxConfig.SignModeHandler().GetSignBytes(
		signMode,
		signerInfo,
		txBuilder.GetTx(),
	)
	suite.Require().NoError(err)

	return signerBytes
}

// createBaseTxBuilder creates a TxBuilder to be used for Single- or Multi-signing
func (suite *AnteTestSuite) createBaseTxBuilder(msg sdk.Msg, gas uint64) client.TxBuilder {
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(gas)
	txBuilder.SetFeeAmount(sdk.NewCoins(
		sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(10000)),
	))

	err := txBuilder.SetMsgs(msg)
	suite.Require().NoError(err)

	txBuilder.SetMemo("")

	return txBuilder
}

// CreateTestSignedMultisigTx creates and sign a multi-signed tx for the given message. `signType` indicates whether to use standard signing ("Standard"),
// EIP-712 signing ("EIP-712"), or a mix of the two ("mixed").
func (suite *AnteTestSuite) CreateTestSignedMultisigTx(privKeys []cryptotypes.PrivKey, signMode signing.SignMode, msg sdk.Msg, chainID string, gas uint64, signType string) client.TxBuilder {
	pubKeys := make([]cryptotypes.PubKey, len(privKeys))
	for i, privKey := range privKeys {
		pubKeys[i] = privKey.PubKey()
	}

	// Re-derive multikey
	numKeys := len(privKeys)
	multiKey := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

	suite.RegisterAccount(multiKey, big.NewInt(10000000000))

	txBuilder := suite.createBaseTxBuilder(msg, gas)

	// Prepare signature field
	sig := multisig.NewMultisig(len(pubKeys))
	err := txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: multiKey,
		Data:   sig,
	})
	suite.Require().NoError(err)

	signerBytes := suite.createSignerBytes(chainID, signMode, multiKey, txBuilder)

	// Sign for each key and update signature field
	sigs := suite.generateMultikeySignatures(signMode, privKeys, signerBytes, signType)
	for _, pkSig := range sigs {
		err := multisig.AddSignatureV2(sig, pkSig, pubKeys)
		suite.Require().NoError(err)
	}

	err = txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: multiKey,
		Data:   sig,
	})
	suite.Require().NoError(err)

	return txBuilder
}

func (suite *AnteTestSuite) CreateTestSingleSignedTx(privKey cryptotypes.PrivKey, signMode signing.SignMode, msg sdk.Msg, chainID string, gas uint64, signType string) client.TxBuilder {
	pubKey := privKey.PubKey()

	suite.RegisterAccount(pubKey, big.NewInt(10000000000))

	txBuilder := suite.createBaseTxBuilder(msg, gas)

	// Prepare signature field
	sig := signing.SingleSignatureData{}
	err := txBuilder.SetSignatures(signing.SignatureV2{
		PubKey: pubKey,
		Data:   &sig,
	})
	suite.Require().NoError(err)

	signerBytes := suite.createSignerBytes(chainID, signMode, pubKey, txBuilder)

	sigData := suite.generateSingleSignature(signMode, privKey, signerBytes, signType)
	err = txBuilder.SetSignatures(sigData)
	suite.Require().NoError(err)

	return txBuilder
}

// BasicSetupForClaimRewardsTest is a helper function, that creates a validator and a delegator
// and funds them with some tokens. It also sets up the staking keeper to include a self-delegation
// of the validator and a delegation from the delegator to the validator.
func (suite *AnteTestSuite) BasicSetupForClaimRewardsTest() (sdk.AccAddress, sdk.ValAddress) {
	// reset historical count
	suite.app.DistrKeeper.DeleteAllValidatorHistoricalRewards(suite.ctx)

	// ----------------------------------------
	// Set up first account
	//
	addr, _ := testutiltx.NewAccAddressAndKey()
	err := testutil.FundAccountWithBaseDenom(suite.ctx, suite.app.BankKeeper, addr, 1e18)
	suite.Require().NoError(err, "failed to fund account")

	// ----------------------------------------
	// Set up validator
	//
	// This account gets funded with the same initial balance as the first account, which
	// will be fully used to self-delegate upon creation of the validator.
	//
	privKey := ed25519.GenPrivKey()
	addr2, _ := testutiltx.NewAccAddressAndKey()
	err = testutil.FundAccountWithBaseDenom(suite.ctx, suite.app.BankKeeper, addr2, 1e18)
	suite.Require().NoError(err, "failed to fund account")

	// 0% decimal for the staking commission
	zeroDec := sdk.ZeroDec()

	// Set up staking parameters
	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = utils.BaseDenom
	stakingParams.MinCommissionRate = zeroDec
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	// Set up staking using helper from sdk testing suite
	stakingHelper := teststaking.NewHelper(suite.T(), suite.ctx, suite.app.StakingKeeper)
	stakingHelper.Commission = stakingtypes.NewCommissionRates(zeroDec, zeroDec, zeroDec)
	stakingHelper.Denom = utils.BaseDenom

	valAddr := sdk.ValAddress(addr2.Bytes())
	stakeAmount := sdk.NewInt(1e10)
	stakingHelper.CreateValidator(valAddr, privKey.PubKey(), stakeAmount, true)
	stakingHelper.Delegate(addr, valAddr, sdk.NewInt(1e10))

	// end block to bond validator and increase block height
	staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	return addr, valAddr
}

// TODO: refactor this together with the method in Cosmos ante handler utils.
// PrepareAccountsForDelegationRewards prepares the test suite for testing to withdraw delegation rewards.
//
// The setup is done in the following way:
//   - Fund the account with the given address with the given balance.
//     If the given balance is zero, the account will be created with zero balance.
//   - Set up a validator with zero commission and delegate to it -> the account delegation will be 50% of the total delegation.
//   - Allocate rewards to the validator.
func PrepareAccountsForDelegationRewards(suite *AnteTestSuite, addr sdk.AccAddress, balance, rewards sdkmath.Int) {
	if balance.IsZero() {
		suite.app.AccountKeeper.SetAccount(suite.ctx, suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr))
	} else {
		// Fund account with enough tokens to stake them
		err := testutil.FundAccountWithBaseDenom(suite.ctx, suite.app.BankKeeper, addr, balance.Int64())
		suite.Require().NoError(err, "failed to fund account")
	}

	if !rewards.IsZero() {
		// reset historical count in distribution keeper which is necessary
		// for the delegation rewards to be calculated correctly
		suite.app.DistrKeeper.DeleteAllValidatorHistoricalRewards(suite.ctx)

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
