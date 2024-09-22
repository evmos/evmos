package evm_test

import (
	"errors"
	"math/big"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethparams "github.com/ethereum/go-ethereum/params"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/x/evm/config"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

func (suite *AnteTestSuite) TestAnteHandler() {
	var (
		ctx     sdk.Context
		addr    common.Address
		privKey cryptotypes.PrivKey
	)
	to := utiltx.GenerateAddress()

	setup := func() {
		suite.WithFeemarketEnabled(false)
		baseFee := sdkmath.NewInt(100)
		suite.WithBaseFee(&baseFee)
		suite.SetupTest() // reset

		fromKey := suite.GetKeyring().GetKey(0)
		addr = fromKey.Addr
		privKey = fromKey.Priv
		ctx = suite.GetNetwork().GetContext()
	}

	ethContractCreationTxParams := evmtypes.EvmTxArgs{
		ChainID:   suite.GetNetwork().App.EvmKeeper.ChainID(),
		Nonce:     0,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasPrice:  big.NewInt(150),
		GasFeeCap: big.NewInt(200),
	}

	ethTxParams := evmtypes.EvmTxArgs{
		ChainID:   suite.GetNetwork().App.EvmKeeper.ChainID(),
		To:        &to,
		Nonce:     0,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasPrice:  big.NewInt(150),
		GasFeeCap: big.NewInt(200),
	}

	baseDenom := config.GetDenom()

	testCases := []struct {
		name      string
		txFn      func() sdk.Tx
		checkTx   bool
		reCheckTx bool
		expPass   bool
	}{
		{
			"success - DeliverTx (contract)",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethContractCreationTxParams)
				suite.Require().NoError(err)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx (contract)",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethContractCreationTxParams)
				suite.Require().NoError(err)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethContractCreationTxParams)
				suite.Require().NoError(err)
				return tx
			},
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethTxParams)
				suite.Require().NoError(err)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethTxParams)
				suite.Require().NoError(err)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethTxParams)
				suite.Require().NoError(err)
				return tx
			}, false, true, true,
		},
		{
			"success - CheckTx (cosmos tx not signed)",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethTxParams)
				suite.Require().NoError(err)
				return tx
			}, false, true, true,
		},
		{
			"fail - CheckTx (cosmos tx is not valid)",
			func() sdk.Tx {
				txBuilder := suite.CreateTxBuilder(privKey, ethTxParams)

				// bigger than MaxGasWanted
				txBuilder.SetGasLimit(uint64(1 << 63))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fail - CheckTx (memo too long)",
			func() sdk.Tx {
				txBuilder := suite.CreateTxBuilder(privKey, ethTxParams)

				txBuilder.SetMemo(strings.Repeat("*", 257))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fail - CheckTx (ExtensionOptionsEthereumTx not set)",
			func() sdk.Tx {
				txBuilder := suite.CreateTxBuilder(privKey, ethTxParams, true)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		// Based on EVMBackend.SendTransaction, for cosmos tx, forcing null for some fields except ExtensionOptions, Fee, MsgEthereumTx
		// should be part of consensus
		{
			"fail - DeliverTx (cosmos tx signed)",
			func() sdk.Tx {
				nonce, err := suite.GetNetwork().App.AccountKeeper.GetSequence(ctx, suite.GetKeyring().GetAccAddr(0))
				suite.Require().NoError(err)
				ethTxParams := evmtypes.EvmTxArgs{
					ChainID:  suite.GetNetwork().App.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}

				txBuilder := suite.CreateTxBuilder(privKey, ethTxParams, true)
				suite.Require().NoError(suite.GetTxFactory().SignCosmosTx(privKey, txBuilder))
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (cosmos tx with memo)",
			func() sdk.Tx {
				nonce, err := suite.GetNetwork().App.AccountKeeper.GetSequence(ctx, suite.GetKeyring().GetAccAddr(0))
				suite.Require().NoError(err)
				ethTxParams := evmtypes.EvmTxArgs{
					ChainID:  suite.GetNetwork().App.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				txBuilder := suite.CreateTxBuilder(privKey, ethTxParams)
				txBuilder.SetMemo("memo for cosmos tx not allowed")
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (cosmos tx with timeoutheight)",
			func() sdk.Tx {
				nonce, err := suite.GetNetwork().App.AccountKeeper.GetSequence(ctx, suite.GetKeyring().GetAccAddr(0))
				suite.Require().NoError(err)
				ethTxParams := evmtypes.EvmTxArgs{
					ChainID:  suite.GetNetwork().App.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				txBuilder := suite.CreateTxBuilder(privKey, ethTxParams)
				txBuilder.SetTimeoutHeight(10)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (invalid fee amount)",
			func() sdk.Tx {
				nonce, err := suite.GetNetwork().App.AccountKeeper.GetSequence(ctx, suite.GetKeyring().GetAccAddr(0))
				suite.Require().NoError(err)
				ethTxParams := evmtypes.EvmTxArgs{
					ChainID:  suite.GetNetwork().App.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				txBuilder := suite.CreateTxBuilder(privKey, ethTxParams)

				expFee := txBuilder.GetTx().GetFee()
				oneCoin := sdk.NewCoin(suite.GetNetwork().GetDenom(), sdkmath.NewInt(1))
				invalidFee := expFee.Add(oneCoin)
				txBuilder.SetFeeAmount(invalidFee)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (invalid fee gaslimit)",
			func() sdk.Tx {
				nonce, err := suite.GetNetwork().App.AccountKeeper.GetSequence(ctx, suite.GetKeyring().GetAccAddr(0))
				suite.Require().NoError(err)
				ethTxParams := evmtypes.EvmTxArgs{
					ChainID:  suite.GetNetwork().App.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				txBuilder := suite.CreateTxBuilder(privKey, ethTxParams)

				expGasLimit := txBuilder.GetTx().GetGas()
				invalidGasLimit := expGasLimit + 1
				txBuilder.SetGasLimit(invalidGasLimit)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx with MsgSend",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(100*int64(gas)))) //#nosec G115
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx with DelegateMsg",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				gas := uint64(200000)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(100*int64(gas))) //#nosec G115
				amount := sdk.NewCoins(coinAmount)
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgDelegate(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 create validator",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgCreateValidator(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 create validator (with blank fields)",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgCreateValidator2(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgSubmitProposal",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				// reusing the gasAmount for deposit
				deposit := sdk.NewCoins(coinAmount)
				txBuilder, err := suite.CreateTestEIP712SubmitProposal(from, privKey, ctx.ChainID(), gas, gasAmount, deposit)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgGrant",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				grantee := sdk.AccAddress("_______grantee______")
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				blockTime := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)
				expiresAt := blockTime.Add(time.Hour)
				msg, err := authz.NewMsgGrant(
					from, grantee, &banktypes.SendAuthorization{SpendLimit: gasAmount}, &expiresAt,
				)
				suite.Require().NoError(err)
				builder, err := suite.CreateTestEIP712SingleMessageTxBuilder(privKey, ctx.ChainID(), gas, gasAmount, msg)
				suite.Require().NoError(err)

				return builder.GetTx()
			}, false, false, true,
		},

		{
			"success- DeliverTx EIP712 MsgGrantAllowance",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712GrantAllowance(from, privKey, ctx.ChainID(), gas, gasAmount)
				suite.Require().NoError(err)

				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 edit validator",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgEditValidator(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 submit evidence",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgSubmitEvidence(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 submit proposal v1",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712SubmitProposalV1(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgExec",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgExec(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgVoteV1",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgVoteV1(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 Multiple MsgSend",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MultipleMsgSend(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 Multiple Different Msgs",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MultipleDifferentMsgs(from, privKey, ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, !suite.useLegacyEIP712TypedData,
		},
		{
			"success- DeliverTx EIP712 Same Msgs, Different Schemas",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712SameMsgDifferentSchemas(from, privKey, ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, !suite.useLegacyEIP712TypedData,
		},
		{
			"success- DeliverTx EIP712 Zero Value Array (Should Not Omit Field)",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712ZeroValueArray(from, privKey, ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, !suite.useLegacyEIP712TypedData,
		},
		{
			"success- DeliverTx EIP712 Zero Value Number (Should Not Omit Field)",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712ZeroValueNumber(from, privKey, ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, !suite.useLegacyEIP712TypedData,
		},
		{
			"success- DeliverTx EIP712 MsgTransfer",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgTransfer(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgTransfer Without Memo",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgTransferWithoutMemo(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"fails - DeliverTx EIP712 Multiple Signers",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				coinAmount := sdk.NewCoin(baseDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MultipleSignerMsgs(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with wrong Chain ID",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(100*int64(gas)))) //#nosec G115
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "evmos_9002-1", gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with different gas fees",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(100*int64(gas)))) //#nosec G115
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				txBuilder.SetGasLimit(uint64(300000))
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(30))))
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid chain id",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(100*int64(gas)))) //#nosec G115
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "evmos_9000-1", gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid sequence",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(100*int64(gas)))) //#nosec G115
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				nonce, err := suite.GetNetwork().App.AccountKeeper.GetSequence(ctx, suite.GetKeyring().GetAccAddr(0))
				suite.Require().NoError(err)
				sigsV2 := signing.SignatureV2{
					PubKey: privKey.PubKey(),
					Data: &signing.SingleSignatureData{
						SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					},
					Sequence: nonce - 1,
				}

				err = txBuilder.SetSignatures(sigsV2)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid signMode",
			func() sdk.Tx {
				from := suite.GetKeyring().GetAccAddr(0)
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(100*int64(gas)))) //#nosec G115
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				nonce, err := suite.GetNetwork().App.AccountKeeper.GetSequence(ctx, suite.GetKeyring().GetAccAddr(0))
				suite.Require().NoError(err)
				sigsV2 := signing.SignatureV2{
					PubKey: privKey.PubKey(),
					Data: &signing.SingleSignatureData{
						SignMode: signing.SignMode_SIGN_MODE_UNSPECIFIED,
					},
					Sequence: nonce,
				}
				err = txBuilder.SetSignatures(sigsV2)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - invalid from",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethContractCreationTxParams)
				suite.Require().NoError(err)
				msg := tx.GetMsgs()[0].(*evmtypes.MsgEthereumTx)
				msg.From = addr.Hex()
				return tx
			}, true, false, false,
		},
		{
			"passes - Single-signer EIP-712",
			func() sdk.Tx {
				msg := banktypes.NewMsgSend(
					sdk.AccAddress(privKey.PubKey().Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"evmos",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSingleSignedTx(
					privKey,
					signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					msg,
					ctx.ChainID(),
					2000000,
					"EIP-712",
				)

				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"passes - EIP-712 multi-key",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"evmos",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					msg,
					ctx.ChainID(),
					2000000,
					"EIP-712",
				)

				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"passes - Mixed multi-key",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"evmos",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					msg,
					ctx.ChainID(),
					2000000,
					"mixed", // Combine EIP-712 and standard signatures
				)

				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"passes - Mixed multi-key with MsgVote",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := govtypes.NewMsgVote(
					sdk.AccAddress(pk.Address()),
					1,
					govtypes.OptionYes,
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					msg,
					ctx.ChainID(),
					2000000,
					"mixed", // Combine EIP-712 and standard signatures
				)

				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"Fails - Multi-Key with incorrect Chain ID",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"evmos",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					msg,
					"evmos_9005-1",
					2000000,
					"mixed",
				)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Multi-Key with incorrect sign mode",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"evmos",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_DIRECT,
					msg,
					ctx.ChainID(),
					2000000,
					"mixed",
				)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Multi-Key with too little gas",
			func() sdk.Tx {
				numKeys := 5
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"evmos",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_DIRECT,
					msg,
					ctx.ChainID(),
					2000,
					"mixed", // Combine EIP-712 and standard signatures
				)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Multi-Key with different payload than one signed",
			func() sdk.Tx {
				numKeys := 1
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"evmos",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_DIRECT,
					msg,
					ctx.ChainID(),
					2000,
					"EIP-712",
				)

				msg.Amount[0].Amount = sdkmath.NewInt(5)
				err := txBuilder.SetMsgs(msg)
				suite.Require().NoError(err)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Multi-Key with messages added after signing",
			func() sdk.Tx {
				numKeys := 1
				privKeys, pubKeys := suite.GenerateMultipleKeys(numKeys)
				pk := kmultisig.NewLegacyAminoPubKey(numKeys, pubKeys)

				msg := banktypes.NewMsgSend(
					sdk.AccAddress(pk.Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"evmos",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSignedMultisigTx(
					privKeys,
					signing.SignMode_SIGN_MODE_DIRECT,
					msg,
					ctx.ChainID(),
					2000,
					"EIP-712",
				)

				// Duplicate
				err := txBuilder.SetMsgs(msg, msg)
				suite.Require().NoError(err)

				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"Fails - Single-Signer EIP-712 with messages added after signing",
			func() sdk.Tx {
				msg := banktypes.NewMsgSend(
					sdk.AccAddress(privKey.PubKey().Address()),
					addr[:],
					sdk.NewCoins(
						sdk.NewCoin(
							"evmos",
							sdkmath.NewInt(1),
						),
					),
				)

				txBuilder := suite.CreateTestSingleSignedTx(
					privKey,
					signing.SignMode_SIGN_MODE_DIRECT,
					msg,
					ctx.ChainID(),
					2000,
					"EIP-712",
				)

				err := txBuilder.SetMsgs(msg, msg)
				suite.Require().NoError(err)

				return txBuilder.GetTx()
			}, false, false, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			setup()

			ctx = ctx.WithIsCheckTx(tc.checkTx).WithIsReCheckTx(tc.reCheckTx)
			anteHandler := suite.GetAnteHandler()
			_, err := anteHandler(ctx, tc.txFn(), false)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AnteTestSuite) TestAnteHandlerWithDynamicTxFee() {
	addr, privKey := utiltx.NewAddrKey()
	to := utiltx.GenerateAddress()

	ethContractCreationTxParams := evmtypes.EvmTxArgs{
		ChainID:   suite.GetNetwork().App.EvmKeeper.ChainID(),
		Nonce:     0,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Accesses:  &types.AccessList{},
	}

	ethTxParams := evmtypes.EvmTxArgs{
		ChainID:   suite.GetNetwork().App.EvmKeeper.ChainID(),
		Nonce:     0,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Accesses:  &types.AccessList{},
		To:        &to,
	}

	testCases := []struct {
		name           string
		txFn           func() sdk.Tx
		enableLondonHF bool
		checkTx        bool
		reCheckTx      bool
		expPass        bool
	}{
		{
			"success - DeliverTx (contract)",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethContractCreationTxParams)
				suite.Require().NoError(err)
				return tx
			},
			true,
			false, false, true,
		},
		{
			"success - CheckTx (contract)",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethContractCreationTxParams)
				suite.Require().NoError(err)
				return tx
			},
			true,
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethContractCreationTxParams)
				suite.Require().NoError(err)
				return tx
			},
			true,
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethTxParams)
				suite.Require().NoError(err)
				return tx
			},
			true,
			false, false, true,
		},
		{
			"success - CheckTx",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethTxParams)
				suite.Require().NoError(err)
				return tx
			},
			true,
			true, false, true,
		},
		{
			"success - ReCheckTx",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethTxParams)
				suite.Require().NoError(err)
				return tx
			},
			true,
			false, true, true,
		},
		{
			"success - CheckTx (cosmos tx not signed)",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethTxParams)
				suite.Require().NoError(err)
				return tx
			},
			true,
			false, true, true,
		},
		{
			"fail - CheckTx (cosmos tx is not valid)",
			func() sdk.Tx {
				txBuilder := suite.CreateTxBuilder(privKey, ethTxParams)
				// bigger than MaxGasWanted
				txBuilder.SetGasLimit(uint64(1 << 63))
				return txBuilder.GetTx()
			},
			true,
			true, false, false,
		},
		{
			"fail - CheckTx (memo too long)",
			func() sdk.Tx {
				txBuilder := suite.CreateTxBuilder(privKey, ethTxParams)
				txBuilder.SetMemo(strings.Repeat("*", 257))
				return txBuilder.GetTx()
			},
			true,
			true, false, false,
		},
		{
			"fail - DynamicFeeTx without london hark fork",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethContractCreationTxParams)
				suite.Require().NoError(err)
				return tx
			},
			false,
			false, false, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.WithFeemarketEnabled(true)
			suite.WithLondonHardForkEnabled(tc.enableLondonHF)
			suite.SetupTest() // reset
			ctx := suite.GetNetwork().GetContext()
			acc := suite.GetNetwork().App.AccountKeeper.NewAccountWithAddress(ctx, addr.Bytes())
			suite.Require().NoError(acc.SetSequence(1))
			suite.GetNetwork().App.AccountKeeper.SetAccount(ctx, acc)

			ctx = ctx.WithIsCheckTx(tc.checkTx).WithIsReCheckTx(tc.reCheckTx)
			err := suite.GetNetwork().App.EvmKeeper.SetBalance(ctx, addr, big.NewInt((ethparams.InitialBaseFee+10)*100000))
			suite.Require().NoError(err)

			anteHandler := suite.GetAnteHandler()
			_, err = anteHandler(ctx, tc.txFn(), false)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.WithFeemarketEnabled(false)
	suite.WithLondonHardForkEnabled(true)
}

func (suite *AnteTestSuite) TestAnteHandlerWithParams() {
	addr, privKey := utiltx.NewAddrKey()
	to := utiltx.GenerateAddress()

	ethContractCreationTxParams := evmtypes.EvmTxArgs{
		ChainID:   suite.GetNetwork().App.EvmKeeper.ChainID(),
		Nonce:     0,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Input:     []byte("create bytes"),
		Accesses:  &types.AccessList{},
	}

	ethTxParams := evmtypes.EvmTxArgs{
		ChainID:   suite.GetNetwork().App.EvmKeeper.ChainID(),
		Nonce:     0,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Accesses:  &types.AccessList{},
		Input:     []byte("call bytes"),
		To:        &to,
	}

	testCases := []struct {
		name        string
		txFn        func() sdk.Tx
		permissions evmtypes.AccessControl
		expErr      error
	}{
		{
			"fail - Contract Creation Disabled",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethContractCreationTxParams)
				suite.Require().NoError(err)
				return tx
			},
			evmtypes.AccessControl{
				Create: evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypeRestricted,
					AccessControlList: evmtypes.DefaultCreateAllowlistAddresses,
				},
				Call: evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypePermissionless,
					AccessControlList: evmtypes.DefaultCreateAllowlistAddresses,
				},
			},
			evmtypes.ErrCreateDisabled,
		},
		{
			"success - Contract Creation Enabled",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethContractCreationTxParams)
				suite.Require().NoError(err)
				return tx
			},
			evmtypes.DefaultAccessControl,
			nil,
		},
		{
			"fail - EVM Call Disabled",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethTxParams)
				suite.Require().NoError(err)
				return tx
			},
			evmtypes.AccessControl{
				Create: evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypePermissionless,
					AccessControlList: evmtypes.DefaultCreateAllowlistAddresses,
				},
				Call: evmtypes.AccessControlType{
					AccessType:        evmtypes.AccessTypeRestricted,
					AccessControlList: evmtypes.DefaultCreateAllowlistAddresses,
				},
			},
			evmtypes.ErrCallDisabled,
		},
		{
			"success - EVM Call Enabled",
			func() sdk.Tx {
				tx, err := suite.GetTxFactory().GenerateSignedEthTx(privKey, ethTxParams)
				suite.Require().NoError(err)
				return tx
			},
			evmtypes.DefaultAccessControl,
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.WithEvmParamsOptions(func(params *evmtypes.Params) {
				params.AccessControl = tc.permissions
			})
			// clean up the evmParamsOption
			defer suite.ResetEvmParamsOptions()

			suite.SetupTest() // reset

			ctx := suite.GetNetwork().GetContext()
			acc := suite.GetNetwork().App.AccountKeeper.NewAccountWithAddress(ctx, addr.Bytes())
			suite.Require().NoError(acc.SetSequence(1))
			suite.GetNetwork().App.AccountKeeper.SetAccount(ctx, acc)

			ctx = ctx.WithIsCheckTx(true)
			err := suite.GetNetwork().App.EvmKeeper.SetBalance(ctx, addr, big.NewInt((ethparams.InitialBaseFee+10)*100000))
			suite.Require().NoError(err)

			anteHandler := suite.GetAnteHandler()
			_, err = anteHandler(ctx, tc.txFn(), false)
			if tc.expErr == nil {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
	suite.WithEvmParamsOptions(nil)
}
