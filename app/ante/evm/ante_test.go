package evm_test

import (
	"errors"
	"math/big"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/ethereum/go-ethereum/core/types"
	ethparams "github.com/ethereum/go-ethereum/params"
	utiltx "github.com/evmos/evmos/v17/testutil/tx"
	evmtypes "github.com/evmos/evmos/v17/x/evm/types"
)

func (suite *AnteTestSuite) TestAnteHandler() {
	var acc authtypes.AccountI
	addr, privKey := utiltx.NewAddrKey()
	to := utiltx.GenerateAddress()

	setup := func() {
		suite.enableFeemarket = false
		suite.SetupTest() // reset

		acc = suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
		suite.Require().NoError(acc.SetSequence(1))
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

		err := suite.app.EvmKeeper.SetBalance(suite.ctx, addr, big.NewInt(10000000000))
		suite.Require().NoError(err)

		suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, big.NewInt(100))
	}

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasPrice:  big.NewInt(150),
		GasFeeCap: big.NewInt(200),
	}

	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		To:        &to,
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasPrice:  big.NewInt(150),
		GasFeeCap: big.NewInt(200),
	}

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
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			}, false, true, true,
		},
		{
			"success - CheckTx (cosmos tx not signed)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			}, false, true, true,
		},
		{
			"fail - CheckTx (cosmos tx is not valid)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				// bigger than MaxGasWanted
				txBuilder.SetGasLimit(uint64(1 << 63))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fail - CheckTx (memo too long)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetMemo(strings.Repeat("*", 257))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fail - CheckTx (ExtensionOptionsEthereumTx not set)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false, true)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		// Based on EVMBackend.SendTransaction, for cosmos tx, forcing null for some fields except ExtensionOptions, Fee, MsgEthereumTx
		// should be part of consensus
		{
			"fail - DeliverTx (cosmos tx signed)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:  suite.app.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, true)
				return tx
			}, false, false, false,
		},
		{
			"fail - DeliverTx (cosmos tx with memo)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:  suite.app.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetMemo("memo for cosmos tx not allowed")
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (cosmos tx with timeoutheight)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:  suite.app.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetTimeoutHeight(10)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (invalid fee amount)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:  suite.app.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)

				txData, err := evmtypes.UnpackTxData(signedTx.Data)
				suite.Require().NoError(err)

				expFee := txData.Fee()
				invalidFee := new(big.Int).Add(expFee, big.NewInt(1))
				invalidFeeAmount := sdk.Coins{sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewIntFromBigInt(invalidFee))}
				txBuilder.SetFeeAmount(invalidFeeAmount)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (invalid fee gaslimit)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:  suite.app.EvmKeeper.ChainID(),
					To:       &to,
					Nonce:    nonce,
					Amount:   big.NewInt(10),
					GasLimit: 100000,
					GasPrice: big.NewInt(1),
				}
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)

				expGasLimit := signedTx.GetGas()
				invalidGasLimit := expGasLimit + 1
				txBuilder.SetGasLimit(invalidGasLimit)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx with MsgSend",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx with DelegateMsg",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas)))
				amount := sdk.NewCoins(coinAmount)
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgDelegate(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 create validator",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgCreateValidator(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 create validator (with blank fields)",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgCreateValidator2(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgSubmitProposal",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				// reusing the gasAmount for deposit
				deposit := sdk.NewCoins(coinAmount)
				txBuilder, err := suite.CreateTestEIP712SubmitProposal(from, privKey, suite.ctx.ChainID(), gas, gasAmount, deposit)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgGrant",
			func() sdk.Tx {
				from := acc.GetAddress()
				grantee := sdk.AccAddress("_______grantee______")
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				blockTime := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)
				expiresAt := blockTime.Add(time.Hour)
				msg, err := authz.NewMsgGrant(
					from, grantee, &banktypes.SendAuthorization{SpendLimit: gasAmount}, &expiresAt,
				)
				suite.Require().NoError(err)
				builder, err := suite.CreateTestEIP712SingleMessageTxBuilder(privKey, suite.ctx.ChainID(), gas, gasAmount, msg)
				suite.Require().NoError(err)

				return builder.GetTx()
			}, false, false, true,
		},

		{
			"success- DeliverTx EIP712 MsgGrantAllowance",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712GrantAllowance(from, privKey, suite.ctx.ChainID(), gas, gasAmount)
				suite.Require().NoError(err)

				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 edit validator",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgEditValidator(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 submit evidence",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgSubmitEvidence(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 submit proposal v1",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712SubmitProposalV1(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgExec",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgExec(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgVoteV1",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgVoteV1(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 Multiple MsgSend",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MultipleMsgSend(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 Multiple Different Msgs",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MultipleDifferentMsgs(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, !suite.useLegacyEIP712TypedData,
		},
		{
			"success- DeliverTx EIP712 Same Msgs, Different Schemas",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712SameMsgDifferentSchemas(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, !suite.useLegacyEIP712TypedData,
		},
		{
			"success- DeliverTx EIP712 Zero Value Array (Should Not Omit Field)",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712ZeroValueArray(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, !suite.useLegacyEIP712TypedData,
		},
		{
			"success- DeliverTx EIP712 Zero Value Number (Should Not Omit Field)",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712ZeroValueNumber(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.RequireErrorForLegacyTypedData(err)
				return suite.TxForLegacyTypedData(txBuilder)
			}, false, false, !suite.useLegacyEIP712TypedData,
		},
		{
			"success- DeliverTx EIP712 MsgTransfer",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgTransfer(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgTransfer Without Memo",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MsgTransferWithoutMemo(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"fails - DeliverTx EIP712 Multiple Signers",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder, err := suite.CreateTestEIP712MultipleSignerMsgs(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with wrong Chain ID",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "evmos_9002-1", gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with different gas fees",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				txBuilder.SetGasLimit(uint64(300000))
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(30))))
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid chain id",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "evmos_9001-1", gas, amount)
				suite.Require().NoError(err)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid sequence",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
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
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder, err := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, suite.ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
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
				msg := evmtypes.NewTx(ethContractCreationTxParams)
				msg.From = addr.Hex()
				tx := suite.CreateTestTx(msg, privKey, 1, false)
				msg = tx.GetMsgs()[0].(*evmtypes.MsgEthereumTx)
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
					suite.ctx.ChainID(),
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
					suite.ctx.ChainID(),
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
					suite.ctx.ChainID(),
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
					suite.ctx.ChainID(),
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
					suite.ctx.ChainID(),
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
					suite.ctx.ChainID(),
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
					suite.ctx.ChainID(),
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
					suite.ctx.ChainID(),
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
					suite.ctx.ChainID(),
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

			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx).WithIsReCheckTx(tc.reCheckTx)

			// expConsumed := params.TxGasContractCreation + params.TxGas
			_, err := suite.anteHandler(suite.ctx, tc.txFn(), false)

			// suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
				// suite.Require().Equal(int(expConsumed), int(suite.ctx.GasMeter().GasConsumed()))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AnteTestSuite) TestAnteHandlerWithDynamicTxFee() {
	addr, privKey := utiltx.NewAddrKey()
	to := utiltx.GenerateAddress()

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Accesses:  &types.AccessList{},
	}

	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
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
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true,
			false, false, true,
		},
		{
			"success - CheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true,
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true,
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true,
			false, false, true,
		},
		{
			"success - CheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true,
			true, false, true,
		},
		{
			"success - ReCheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true,
			false, true, true,
		},
		{
			"success - CheckTx (cosmos tx not signed)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true,
			false, true, true,
		},
		{
			"fail - CheckTx (cosmos tx is not valid)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
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
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetMemo(strings.Repeat("*", 257))
				return txBuilder.GetTx()
			},
			true,
			true, false, false,
		},
		{
			"fail - DynamicFeeTx without london hark fork",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			false,
			false, false, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.enableFeemarket = true
			suite.enableLondonHF = tc.enableLondonHF
			suite.SetupTest() // reset

			acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
			suite.Require().NoError(acc.SetSequence(1))
			suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx).WithIsReCheckTx(tc.reCheckTx)
			err := suite.app.EvmKeeper.SetBalance(suite.ctx, addr, big.NewInt((ethparams.InitialBaseFee+10)*100000))
			suite.Require().NoError(err)

			_, err = suite.anteHandler(suite.ctx, tc.txFn(), false)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.enableFeemarket = false
	suite.enableLondonHF = true
}

func (suite *AnteTestSuite) TestAnteHandlerWithParams() {
	addr, privKey := utiltx.NewAddrKey()
	to := utiltx.GenerateAddress()

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Input:     []byte("create bytes"),
		Accesses:  &types.AccessList{},
	}

	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(ethparams.InitialBaseFee + 1),
		GasTipCap: big.NewInt(1),
		Accesses:  &types.AccessList{},
		Input:     []byte("call bytes"),
		To:        &to,
	}

	testCases := []struct {
		name         string
		txFn         func() sdk.Tx
		enableCall   bool
		enableCreate bool
		expErr       error
	}{
		{
			"fail - Contract Creation Disabled",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true, false,
			evmtypes.ErrCreateDisabled,
		},
		{
			"success - Contract Creation Enabled",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTx(ethContractCreationTxParams)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true, true,
			nil,
		},
		{
			"fail - EVM Call Disabled",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			false, true,
			evmtypes.ErrCallDisabled,
		},
		{
			"success - EVM Call Enabled",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(ethTxParams)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true, true,
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.evmParamsOption = func(params *evmtypes.Params) {
				params.EnableCall = tc.enableCall
				params.EnableCreate = tc.enableCreate
			}
			suite.SetupTest() // reset

			acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
			suite.Require().NoError(acc.SetSequence(1))
			suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

			suite.ctx = suite.ctx.WithIsCheckTx(true)
			err := suite.app.EvmKeeper.SetBalance(suite.ctx, addr, big.NewInt((ethparams.InitialBaseFee+10)*100000))
			suite.Require().NoError(err)

			_, err = suite.anteHandler(suite.ctx, tc.txFn(), false)
			if tc.expErr == nil {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
	suite.evmParamsOption = nil
}
