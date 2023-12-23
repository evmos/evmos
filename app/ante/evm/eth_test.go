package evm_test

import ()

// func (suite *AnteTestSuite) TestEthGasConsumeDecorator() {
// 	chainID := suite.app.EvmKeeper.ChainID()
// 	dec := ethante.NewEthGasConsumeDecorator(suite.app.BankKeeper, suite.app.DistrKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper, config.DefaultMaxTxGasWanted)
//
// 	addr := testutiltx.GenerateAddress()
//
// 	txGasLimit := uint64(1000)
//
// 	ethContractCreationTxParams := &evmtypes.EvmTxArgs{
// 		ChainID:  chainID,
// 		Nonce:    1,
// 		Amount:   big.NewInt(10),
// 		GasLimit: txGasLimit,
// 		GasPrice: big.NewInt(1),
// 	}
//
// 	tx := evmtypes.NewTx(ethContractCreationTxParams)
// 	tx.From = addr.Hex()
//
// 	ethContractCreationTxParams.GasLimit = 55000
// 	zeroFeeTx := makeZeroFeeTx(addr, *ethContractCreationTxParams)
//
// 	ethCfg := suite.app.EvmKeeper.GetParams(suite.ctx).
// 		ChainConfig.EthereumConfig(chainID)
// 	baseFee := suite.app.EvmKeeper.GetBaseFee(suite.ctx, ethCfg)
// 	suite.Require().Equal(int64(1000000000), baseFee.Int64())
//
// 	gasPrice := new(big.Int).Add(baseFee, evmtypes.DefaultPriorityReduction.BigInt())
//
// 	tx2GasLimit := uint64(1000000)
// 	eth2TxContractParams := &evmtypes.EvmTxArgs{
// 		ChainID:  chainID,
// 		Nonce:    1,
// 		Amount:   big.NewInt(10),
// 		GasLimit: tx2GasLimit,
// 		GasPrice: gasPrice,
// 		Accesses: &ethtypes.AccessList{{Address: addr, StorageKeys: nil}},
// 	}
// 	tx2 := evmtypes.NewTx(eth2TxContractParams)
// 	tx2.From = addr.Hex()
// 	tx2Priority := int64(1)
//
// 	tx3GasLimit := types.BlockGasLimit(suite.ctx) + uint64(1)
// 	eth3TxContractParams := &evmtypes.EvmTxArgs{
// 		ChainID:  chainID,
// 		Nonce:    1,
// 		Amount:   big.NewInt(10),
// 		GasLimit: tx3GasLimit,
// 		GasPrice: gasPrice,
// 		Accesses: &ethtypes.AccessList{{Address: addr, StorageKeys: nil}},
// 	}
// 	tx3 := evmtypes.NewTx(eth3TxContractParams)
//
// 	dynamicTxContractParams := &evmtypes.EvmTxArgs{
// 		ChainID:   chainID,
// 		Nonce:     1,
// 		Amount:    big.NewInt(10),
// 		GasLimit:  tx2GasLimit,
// 		GasFeeCap: new(big.Int).Add(baseFee, big.NewInt(evmtypes.DefaultPriorityReduction.Int64()*2)),
// 		GasTipCap: evmtypes.DefaultPriorityReduction.BigInt(),
// 		Accesses:  &ethtypes.AccessList{{Address: addr, StorageKeys: nil}},
// 	}
// 	dynamicFeeTx := evmtypes.NewTx(dynamicTxContractParams)
// 	dynamicFeeTx.From = addr.Hex()
// 	dynamicFeeTxPriority := int64(1)
//
// 	var vmdb *statedb.StateDB
//
// 	initialBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr.Bytes(), utils.BaseDenom)
//
// 	testCases := []struct {
// 		name        string
// 		tx          sdk.Tx
// 		gasLimit    uint64
// 		malleate    func(ctx sdk.Context) sdk.Context
// 		expPass     bool
// 		expPanic    bool
// 		expPriority int64
// 		postCheck   func(ctx sdk.Context)
// 	}{
// 		{
// 			"invalid transaction type",
// 			&testutiltx.InvalidTx{},
// 			math.MaxUint64,
// 			func(ctx sdk.Context) sdk.Context { return ctx },
// 			false,
// 			false,
// 			0,
// 			func(ctx sdk.Context) {},
// 		},
// 		{
// 			"sender not found",
// 			evmtypes.NewTx(&evmtypes.EvmTxArgs{
// 				ChainID:  chainID,
// 				Nonce:    1,
// 				Amount:   big.NewInt(10),
// 				GasLimit: 1000,
// 				GasPrice: big.NewInt(1),
// 			}),
// 			math.MaxUint64,
// 			func(ctx sdk.Context) sdk.Context { return ctx },
// 			false, false,
// 			0,
// 			func(ctx sdk.Context) {},
// 		},
// 		{
// 			"gas limit too low",
// 			tx,
// 			math.MaxUint64,
// 			func(ctx sdk.Context) sdk.Context { return ctx },
// 			false, false,
// 			0,
// 			func(ctx sdk.Context) {},
// 		},
// 		{
// 			"gas limit above block gas limit",
// 			tx3,
// 			math.MaxUint64,
// 			func(ctx sdk.Context) sdk.Context { return ctx },
// 			false, false,
// 			0,
// 			func(ctx sdk.Context) {},
// 		},
// 		{
// 			"not enough balance for fees",
// 			tx2,
// 			math.MaxUint64,
// 			func(ctx sdk.Context) sdk.Context { return ctx },
// 			false, false,
// 			0,
// 			func(ctx sdk.Context) {},
// 		},
// 		{
// 			"not enough tx gas",
// 			tx2,
// 			0,
// 			func(ctx sdk.Context) sdk.Context {
// 				vmdb.AddBalance(addr, big.NewInt(1e6))
// 				return ctx
// 			},
// 			false, true,
// 			0,
// 			func(ctx sdk.Context) {},
// 		},
// 		{
// 			"not enough block gas",
// 			tx2,
// 			0,
// 			func(ctx sdk.Context) sdk.Context {
// 				vmdb.AddBalance(addr, big.NewInt(1e6))
// 				return ctx.WithBlockGasMeter(storetypes.NewGasMeter(1))
// 			},
// 			false, true,
// 			0,
// 			func(ctx sdk.Context) {},
// 		},
// 		{
// 			"success - legacy tx",
// 			tx2,
// 			tx2GasLimit, // it's capped
// 			func(ctx sdk.Context) sdk.Context {
// 				vmdb.AddBalance(addr, big.NewInt(1e16))
// 				return ctx.WithBlockGasMeter(storetypes.NewGasMeter(1e19))
// 			},
// 			true, false,
// 			tx2Priority,
// 			func(ctx sdk.Context) {},
// 		},
// 		{
// 			"success - dynamic fee tx",
// 			dynamicFeeTx,
// 			tx2GasLimit, // it's capped
// 			func(ctx sdk.Context) sdk.Context {
// 				vmdb.AddBalance(addr, big.NewInt(1e16))
// 				return ctx.WithBlockGasMeter(storetypes.NewGasMeter(1e19))
// 			},
// 			true, false,
// 			dynamicFeeTxPriority,
// 			func(ctx sdk.Context) {},
// 		},
// 		{
// 			"success - gas limit on gasMeter is set on ReCheckTx mode",
// 			dynamicFeeTx,
// 			0, // for reCheckTX mode, gas limit should be set to 0
// 			func(ctx sdk.Context) sdk.Context {
// 				vmdb.AddBalance(addr, big.NewInt(1e15))
// 				return ctx.WithIsReCheckTx(true)
// 			},
// 			true, false,
// 			0,
// 			func(ctx sdk.Context) {},
// 		},
// 		{
// 			"success - legacy tx - insufficient funds but enough staking rewards",
// 			tx2,
// 			tx2GasLimit, // it's capped
// 			func(ctx sdk.Context) sdk.Context {
// 				return suite.prepareAccount(
// 					ctx,
// 					addr.Bytes(),
// 					sdkmath.ZeroInt(),
// 					sdkmath.NewInt(1e16),
// 				)
// 			},
// 			true, false,
// 			tx2Priority,
// 			func(ctx sdk.Context) {
// 				balance := suite.app.BankKeeper.GetBalance(ctx, sdk.AccAddress(addr.Bytes()), utils.BaseDenom)
// 				suite.Require().False(
// 					balance.Amount.IsZero(),
// 					"the fees are paid after withdrawing (a surplus amount of) staking rewards, so it should be higher than the initial balance",
// 				)
//
// 				rewards, err := testutil.GetTotalDelegationRewards(ctx, suite.app.DistrKeeper, sdk.AccAddress(addr.Bytes()))
// 				suite.Require().NoError(err, "error while querying delegation total rewards")
// 				suite.Require().Nil(rewards, "the total rewards should be nil after withdrawing all of them")
// 			},
// 		},
// 		{
// 			"success - legacy tx - enough funds so no staking rewards should be used",
// 			tx2,
// 			tx2GasLimit, // it's capped
// 			func(ctx sdk.Context) sdk.Context {
// 				return suite.prepareAccount(
// 					ctx,
// 					addr.Bytes(),
// 					sdkmath.NewInt(1e16),
// 					sdkmath.NewInt(1e16),
// 				)
// 			},
// 			true, false,
// 			tx2Priority,
// 			func(ctx sdk.Context) {
// 				balance := suite.app.BankKeeper.GetBalance(ctx, sdk.AccAddress(addr.Bytes()), utils.BaseDenom)
// 				suite.Require().True(
// 					balance.Amount.LT(sdkmath.NewInt(1e16)),
// 					"the fees are paid using the available balance, so it should be lower than the initial balance",
// 				)
//
// 				rewards, err := testutil.GetTotalDelegationRewards(ctx, suite.app.DistrKeeper, sdk.AccAddress(addr.Bytes()))
// 				suite.Require().NoError(err, "error while querying delegation total rewards")
//
// 				// NOTE: the total rewards should be the same as after the setup, since
// 				// the fees are paid using the account balance
// 				suite.Require().Equal(
// 					sdk.NewDecCoins(sdk.NewDecCoin(utils.BaseDenom, sdkmath.NewInt(1e16))),
// 					rewards,
// 					"the total rewards should be the same as after the setup, since the fees are paid using the account balance",
// 				)
// 			},
// 		},
// 		{
// 			"success - zero fees (no base fee)",
// 			zeroFeeTx,
// 			zeroFeeTx.GetGas(),
// 			func(ctx sdk.Context) sdk.Context {
// 				suite.disableBaseFee(ctx)
// 				return ctx
// 			},
// 			true, false,
// 			0,
// 			func(ctx sdk.Context) {
// 				finalBalance := suite.app.BankKeeper.GetBalance(ctx, addr.Bytes(), utils.BaseDenom)
// 				suite.Require().Equal(initialBalance, finalBalance)
// 			},
// 		},
// 		{
// 			"success - zero fees (no base fee) - legacy tx",
// 			makeZeroFeeTx(addr, *eth2TxContractParams),
// 			tx2GasLimit,
// 			func(ctx sdk.Context) sdk.Context {
// 				suite.disableBaseFee(ctx)
// 				return ctx
// 			},
// 			true, false,
// 			0,
// 			func(ctx sdk.Context) {
// 				finalBalance := suite.app.BankKeeper.GetBalance(ctx, addr.Bytes(), utils.BaseDenom)
// 				suite.Require().Equal(initialBalance, finalBalance)
// 			},
// 		},
// 	}
//
// 	for _, tc := range testCases {
// 		suite.Run(tc.name, func() {
// 			cacheCtx, _ := suite.ctx.CacheContext()
// 			// Create new stateDB for each test case from the cached context
// 			vmdb = testutil.NewStateDB(cacheCtx, suite.app.EvmKeeper)
// 			cacheCtx = tc.malleate(cacheCtx)
// 			suite.Require().NoError(vmdb.Commit())
//
// 			if tc.expPanic {
// 				suite.Require().Panics(func() {
// 					_, _ = dec.AnteHandle(cacheCtx.WithIsCheckTx(true).WithGasMeter(storetypes.NewGasMeter(1)), tc.tx, false, testutil.NextFn)
// 				})
// 				return
// 			}
//
// 			ctx, err := dec.AnteHandle(cacheCtx.WithIsCheckTx(true).WithGasMeter(sdk.NewInfiniteGasMeter()), tc.tx, false, testutil.NextFn)
// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(tc.expPriority, ctx.Priority())
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 			suite.Require().Equal(tc.gasLimit, ctx.GasMeter().Limit())
//
// 			// check state after the test case
// 			tc.postCheck(ctx)
// 		})
// 	}
// }
