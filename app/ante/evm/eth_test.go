package evm_test

import (
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethante "github.com/evmos/evmos/v11/app/ante/evm"
	"github.com/evmos/evmos/v11/server/config"
	"github.com/evmos/evmos/v11/testutil"
	testutiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/types"
	"github.com/evmos/evmos/v11/utils"
	"github.com/evmos/evmos/v11/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v11/x/evm/types"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func (suite *AnteTestSuite) TestNewEthAccountVerificationDecorator() {
	dec := ethante.NewEthAccountVerificationDecorator(
		suite.app.AccountKeeper, suite.app.EvmKeeper,
	)

	addr := testutiltx.GenerateAddress()

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}

	tx := evmtypes.NewTx(ethContractCreationTxParams)
	tx.From = addr.Hex()

	var vmdb *statedb.StateDB

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		checkTx  bool
		expPass  bool
	}{
		{"not CheckTx", nil, func() {}, false, true},
		{"invalid transaction type", &testutiltx.InvalidTx{}, func() {}, true, false},
		{
			"sender not set to msg",
			tx,
			func() {},
			true,
			false,
		},
		{
			"sender not EOA",
			tx,
			func() {
				// set not as an EOA
				vmdb.SetCode(addr, []byte("1"))
			},
			true,
			false,
		},
		{
			"not enough balance to cover tx cost",
			tx,
			func() {
				// reset back to EOA
				vmdb.SetCode(addr, nil)
			},
			true,
			false,
		},
		{
			"success new account",
			tx,
			func() {
				vmdb.AddBalance(addr, big.NewInt(1000000))
			},
			true,
			true,
		},
		{
			"success existing account",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				vmdb.AddBalance(addr, big.NewInt(1000000))
			},
			true,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			vmdb = suite.StateDB()
			tc.malleate()
			suite.Require().NoError(vmdb.Commit())

			_, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(tc.checkTx), tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AnteTestSuite) TestEthNonceVerificationDecorator() {
	suite.SetupTest()
	dec := ethante.NewEthIncrementSenderSequenceDecorator(suite.app.AccountKeeper)

	addr := testutiltx.GenerateAddress()

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}

	tx := evmtypes.NewTx(ethContractCreationTxParams)
	tx.From = addr.Hex()

	testCases := []struct {
		name      string
		tx        sdk.Tx
		malleate  func()
		reCheckTx bool
		expPass   bool
	}{
		{"ReCheckTx", &testutiltx.InvalidTx{}, func() {}, true, false},
		{"invalid transaction type", &testutiltx.InvalidTx{}, func() {}, false, false},
		{"sender account not found", tx, func() {}, false, false},
		{
			"sender nonce missmatch",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false,
			false,
		},
		{
			"success",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.Require().NoError(acc.SetSequence(1))
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()
			_, err := dec.AnteHandle(suite.ctx.WithIsReCheckTx(tc.reCheckTx), tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AnteTestSuite) TestEthGasConsumeDecorator() {
	chainID := suite.app.EvmKeeper.ChainID()
	dec := ethante.NewEthGasConsumeDecorator(suite.app.BankKeeper, suite.app.DistrKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper, config.DefaultMaxTxGasWanted)

	addr := testutiltx.GenerateAddress()

	txGasLimit := uint64(1000)

	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:  chainID,
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: txGasLimit,
		GasPrice: big.NewInt(1),
	}

	tx := evmtypes.NewTx(ethContractCreationTxParams)
	tx.From = addr.Hex()

	ethCfg := suite.app.EvmKeeper.GetParams(suite.ctx).
		ChainConfig.EthereumConfig(chainID)
	baseFee := suite.app.EvmKeeper.GetBaseFee(suite.ctx, ethCfg)
	suite.Require().Equal(int64(1000000000), baseFee.Int64())

	gasPrice := new(big.Int).Add(baseFee, evmtypes.DefaultPriorityReduction.BigInt())

	tx2GasLimit := uint64(1000000)
	eth2TxContractParams := &evmtypes.EvmTxArgs{
		ChainID:  chainID,
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: tx2GasLimit,
		GasPrice: gasPrice,
		Accesses: &ethtypes.AccessList{{Address: addr, StorageKeys: nil}},
	}
	tx2 := evmtypes.NewTx(eth2TxContractParams)
	tx2.From = addr.Hex()
	tx2Priority := int64(1)

	tx3GasLimit := types.BlockGasLimit(suite.ctx) + uint64(1)
	eth3TxContractParams := &evmtypes.EvmTxArgs{
		ChainID:  chainID,
		Nonce:    1,
		Amount:   big.NewInt(10),
		GasLimit: tx3GasLimit,
		GasPrice: gasPrice,
		Accesses: &ethtypes.AccessList{{Address: addr, StorageKeys: nil}},
	}
	tx3 := evmtypes.NewTx(eth3TxContractParams)

	dynamicTxContractParams := &evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  tx2GasLimit,
		GasFeeCap: new(big.Int).Add(baseFee, big.NewInt(evmtypes.DefaultPriorityReduction.Int64()*2)),
		GasTipCap: evmtypes.DefaultPriorityReduction.BigInt(),
		Accesses:  &ethtypes.AccessList{{Address: addr, StorageKeys: nil}},
	}
	dynamicFeeTx := evmtypes.NewTx(dynamicTxContractParams)
	dynamicFeeTx.From = addr.Hex()
	dynamicFeeTxPriority := int64(1)

	// store context before testcase execution, so it can always be reset
	defaultCtx := suite.ctx

	var vmdb *statedb.StateDB

	testCases := []struct {
		name        string
		tx          sdk.Tx
		gasLimit    uint64
		malleate    func()
		expPass     bool
		expPanic    bool
		expPriority int64
		postCheck   func()
	}{
		{
			"invalid transaction type",
			&testutiltx.InvalidTx{},
			math.MaxUint64,
			func() {},
			false,
			false,
			0,
			func() {},
		},
		{
			"sender not found",
			evmtypes.NewTx(&evmtypes.EvmTxArgs{
				ChainID:  chainID,
				Nonce:    1,
				Amount:   big.NewInt(10),
				GasLimit: 1000,
				GasPrice: big.NewInt(1),
			}),
			math.MaxUint64,
			func() {},
			false, false,
			0,
			func() {},
		},
		{
			"gas limit too low",
			tx,
			math.MaxUint64,
			func() {},
			false, false,
			0,
			func() {},
		},
		{
			"gas limit above block gas limit",
			tx3,
			math.MaxUint64,
			func() {},
			false, false,
			0,
			func() {},
		},
		{
			"not enough balance for fees",
			tx2,
			math.MaxUint64,
			func() {},
			false, false,
			0,
			func() {},
		},
		{
			"not enough tx gas",
			tx2,
			0,
			func() {
				vmdb.AddBalance(addr, big.NewInt(1e6))
			},
			false, true,
			0,
			func() {},
		},
		{
			"not enough block gas",
			tx2,
			0,
			func() {
				vmdb.AddBalance(addr, big.NewInt(1e6))
				suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(1))
			},
			false, true,
			0,
			func() {},
		},
		{
			"success - legacy tx",
			tx2,
			tx2GasLimit, // it's capped
			func() {
				vmdb.AddBalance(addr, big.NewInt(1e16))
				suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(10000000000000000000))
			},
			true, false,
			tx2Priority,
			func() {},
		},
		{
			"success - dynamic fee tx",
			dynamicFeeTx,
			tx2GasLimit, // it's capped
			func() {
				vmdb.AddBalance(addr, big.NewInt(1e15))
				suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(10000000000000000000))
			},
			true, false,
			dynamicFeeTxPriority,
			func() {},
		},
		{
			"success - gas limit on gasMeter is set on ReCheckTx mode",
			dynamicFeeTx,
			0, // for reCheckTX mode, gas limit should be set to 0
			func() {
				vmdb.AddBalance(addr, big.NewInt(1e15))
				suite.ctx = suite.ctx.WithIsReCheckTx(true)
			},
			true, false,
			0,
			func() {},
		},
		{
			"success - legacy tx - insufficient funds but enough staking rewards",
			tx2,
			tx2GasLimit, // it's capped
			func() {
				vmdb.AddBalance(addr, big.NewInt(1e6))
				suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(1e19))

				PrepareAccountsForDelegationRewards(
					suite, sdk.AccAddress(addr.Bytes()), sdk.NewInt(1e16), sdk.NewInt(1e16),
				)
			},
			true, false,
			tx2Priority,
			func() {
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(addr.Bytes()), utils.BaseDenom)
				suite.Require().True(
					balance.Amount.GT(sdk.NewInt(1e6)),
					"after withdrawing staking rewards, the balance should have increased.",
				)
			},
		},
		{
			"success - legacy tx - enough funds so no staking rewards should be used",
			tx2,
			tx2GasLimit, // it's capped
			func() {
				vmdb.AddBalance(addr, big.NewInt(1e16))
				suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(1e19))

				// NOTE: a certain balance has to be assigned to the account to allow for the delegation
				PrepareAccountsForDelegationRewards(
					suite, sdk.AccAddress(addr.Bytes()), sdk.NewInt(1e16), sdk.NewInt(1e16),
				)
			},
			true, false,
			tx2Priority,
			func() {
				balance := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(addr.Bytes()), utils.BaseDenom)
				suite.Require().True(
					balance.Amount.LT(sdk.NewInt(1e16)),
					"the fees are paid using the available balance, so it should be lower than the initial balance",
				)

				// NOTE: 2e16 is the total rewards because there are 2*1e16 rewards being allocated
				// to the validator to establish a 50/50 split between self-delegation and the account delegation
				suite.Require().Equal(
					sdk.NewDecCoins(sdk.NewDecCoin(utils.BaseDenom, sdk.NewInt(2e16))),
					suite.app.DistrKeeper.GetTotalRewards(suite.ctx),
					"the total rewards should be the same as after the setup, since the fees are paid using the account balance",
				)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			vmdb = suite.StateDB()
			suite.ctx = defaultCtx // reset context before executing testcases
			tc.malleate()
			suite.Require().NoError(vmdb.Commit())

			if tc.expPanic {
				suite.Require().Panics(func() {
					_, _ = dec.AnteHandle(suite.ctx.WithIsCheckTx(true).WithGasMeter(sdk.NewGasMeter(1)), tc.tx, false, testutil.NextFn)
				})
				return
			}

			ctx, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(true).WithGasMeter(sdk.NewInfiniteGasMeter()), tc.tx, false, testutil.NextFn)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expPriority, ctx.Priority())
			} else {
				suite.Require().Error(err)
			}
			suite.Require().Equal(tc.gasLimit, ctx.GasMeter().Limit())

			// check state after the test case
			suite.ctx = ctx
			tc.postCheck()
		})
	}
}

func (suite *AnteTestSuite) TestCanTransferDecorator() {
	dec := ethante.NewCanTransferDecorator(suite.app.EvmKeeper)

	addr, privKey := testutiltx.NewAddrKey()

	suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, big.NewInt(100))
	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
		ChainID:   suite.app.EvmKeeper.ChainID(),
		Nonce:     1,
		Amount:    big.NewInt(10),
		GasLimit:  1000,
		GasPrice:  big.NewInt(1),
		GasFeeCap: big.NewInt(150),
		GasTipCap: big.NewInt(200),
		Accesses:  &ethtypes.AccessList{},
	}

	tx := evmtypes.NewTx(ethContractCreationTxParams)
	tx2 := evmtypes.NewTx(ethContractCreationTxParams)

	tx.From = addr.Hex()

	err := tx.Sign(suite.ethSigner, testutiltx.NewSigner(privKey))
	suite.Require().NoError(err)

	var vmdb *statedb.StateDB

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		expPass  bool
	}{
		{"invalid transaction type", &testutiltx.InvalidTx{}, func() {}, false},
		{"AsMessage failed", tx2, func() {}, false},
		{
			"evm CanTransfer failed",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false,
		},
		{
			"success",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				vmdb.AddBalance(addr, big.NewInt(1000000))
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			vmdb = suite.StateDB()
			tc.malleate()
			suite.Require().NoError(vmdb.Commit())

			_, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(true), tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *AnteTestSuite) TestEthIncrementSenderSequenceDecorator() {
	dec := ethante.NewEthIncrementSenderSequenceDecorator(suite.app.AccountKeeper)
	addr, privKey := testutiltx.NewAddrKey()

	ethTxContractParamsNonce0 := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    0,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}
	contract := evmtypes.NewTx(ethTxContractParamsNonce0)
	contract.From = addr.Hex()
	err := contract.Sign(suite.ethSigner, testutiltx.NewSigner(privKey))
	suite.Require().NoError(err)

	to := testutiltx.GenerateAddress()
	ethTxParamsNonce0 := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    0,
		To:       &to,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}
	tx := evmtypes.NewTx(ethTxParamsNonce0)
	tx.From = addr.Hex()
	err = tx.Sign(suite.ethSigner, testutiltx.NewSigner(privKey))
	suite.Require().NoError(err)

	ethTxParamsNonce1 := &evmtypes.EvmTxArgs{
		ChainID:  suite.app.EvmKeeper.ChainID(),
		Nonce:    1,
		To:       &to,
		Amount:   big.NewInt(10),
		GasLimit: 1000,
		GasPrice: big.NewInt(1),
	}
	tx2 := evmtypes.NewTx(ethTxParamsNonce1)
	tx2.From = addr.Hex()
	err = tx2.Sign(suite.ethSigner, testutiltx.NewSigner(privKey))
	suite.Require().NoError(err)

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		expPass  bool
		expPanic bool
	}{
		{
			"invalid transaction type",
			&testutiltx.InvalidTx{},
			func() {},
			false, false,
		},
		{
			"no signers",
			evmtypes.NewTx(ethTxParamsNonce1),
			func() {},
			false, false,
		},
		{
			"account not set to store",
			tx,
			func() {},
			false, false,
		},
		{
			"success - create contract",
			contract,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			true, false,
		},
		{
			"success - call",
			tx2,
			func() {},
			true, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			if tc.expPanic {
				suite.Require().Panics(func() {
					_, _ = dec.AnteHandle(suite.ctx, tc.tx, false, testutil.NextFn)
				})
				return
			}

			_, err := dec.AnteHandle(suite.ctx, tc.tx, false, testutil.NextFn)

			if tc.expPass {
				suite.Require().NoError(err)
				msg := tc.tx.(*evmtypes.MsgEthereumTx)

				txData, err := evmtypes.UnpackTxData(msg.Data)
				suite.Require().NoError(err)

				nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, addr)
				suite.Require().Equal(txData.GetNonce()+1, nonce)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
