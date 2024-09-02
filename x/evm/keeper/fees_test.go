package keeper_test

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v19/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

func (suite *KeeperTestSuite) TestCheckSenderBalance() {
	hundredInt := sdkmath.NewInt(1000000000000)
	zeroInt := sdkmath.ZeroInt()
	oneInt := sdkmath.NewInt(1000000000)
	fiveInt := sdkmath.NewInt(5000000000)
	fiftyInt := sdkmath.NewInt(50)
	negInt := sdkmath.NewInt(-10)
	oneIntUnscaled := sdkmath.NewInt(1)

	testCases := []struct {
		name            string
		to              string
		gasLimit        uint64
		gasPrice        *sdkmath.Int
		gasFeeCap       *big.Int
		gasTipCap       *big.Int
		cost            *sdkmath.Int
		from            string
		accessList      *ethtypes.AccessList
		expectPass      bool
		enableFeemarket bool
	}{
		{
			name:       "Enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Equal balance",
			to:         suite.address.String(),
			gasLimit:   99,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "negative cost",
			to:         suite.address.String(),
			gasLimit:   1,
			gasPrice:   &oneIntUnscaled,
			cost:       &negInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher gas limit, not enough balance",
			to:         suite.address.String(),
			gasLimit:   10000000000000,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher gas price, enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Higher gas price, not enough balance",
			to:         suite.address.String(),
			gasLimit:   30000000,
			gasPrice:   &fiveInt,
			cost:       &oneInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher cost, enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &fiftyInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Higher cost, not enough balance",
			to:         suite.address.String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &hundredInt,
			from:       suite.address.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:            "Enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(1),
			cost:            &oneInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "Equal balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        99,
			gasFeeCap:       big.NewInt(1),
			cost:            &oneInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "negative cost w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        1,
			gasFeeCap:       big.NewInt(1),
			cost:            &negInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
		{
			name:            "Higher gas limit, not enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        3000000000000,
			gasFeeCap:       big.NewInt(1),
			cost:            &oneInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
		{
			name:            "Higher gas price, enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(5),
			cost:            &oneInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "Higher gas price, not enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        3000000000000,
			gasFeeCap:       big.NewInt(5),
			gasPrice:        &fiveInt,
			cost:            &oneInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
		{
			name:            "Higher cost, enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(5),
			cost:            &fiftyInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "Higher cost, not enough balance w/ enableFeemarket",
			to:              suite.address.String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(5),
			cost:            &hundredInt,
			from:            suite.address.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
	}

	vmdb := suite.StateDB()
	vmdb.AddBalance(suite.address, hundredInt.BigInt())
	balance := vmdb.GetBalance(suite.address)
	suite.Require().Equal(balance, hundredInt.BigInt())
	err := vmdb.Commit()
	suite.Require().NoError(err, "Unexpected error while committing to vmdb: %d", err)

	for i, tc := range testCases {
		suite.Run(tc.name, func() {
			to := common.HexToAddress(tc.from)

			var amount, gasPrice, gasFeeCap, gasTipCap *big.Int
			if tc.cost != nil {
				amount = tc.cost.BigInt()
			}

			if tc.enableFeemarket {
				gasFeeCap = tc.gasFeeCap
				if tc.gasTipCap == nil {
					gasTipCap = oneInt.BigInt()
				} else {
					gasTipCap = tc.gasTipCap
				}
			} else if tc.gasPrice != nil {
				gasPrice = tc.gasPrice.BigInt()
			}

			ethTxParams := &evmtypes.EvmTxArgs{
				ChainID:   zeroInt.BigInt(),
				Nonce:     1,
				To:        &to,
				Amount:    amount,
				GasLimit:  tc.gasLimit,
				GasPrice:  gasPrice,
				GasFeeCap: gasFeeCap,
				GasTipCap: gasTipCap,
				Accesses:  tc.accessList,
			}
			tx := evmtypes.NewTx(ethTxParams)
			tx.From = tc.from

			txData, _ := evmtypes.UnpackTxData(tx.Data)

			acct := suite.app.EvmKeeper.GetAccountOrEmpty(suite.ctx, suite.address)
			err := keeper.CheckSenderBalance(
				sdkmath.NewIntFromBigInt(acct.Balance),
				txData,
			)

			if tc.expectPass {
				suite.Require().NoError(err, "valid test %d failed", i)
			} else {
				suite.Require().Error(err, "invalid test %d passed", i)
			}
		})
	}
}

// TestVerifyFeeAndDeductTxCostsFromUserBalance is a test method for both the VerifyFee
// function and the DeductTxCostsFromUserBalance method.
//
// NOTE: This method combines testing for both functions, because these used to be
// in one function and share a lot of the same setup.
// In practice, the two tested functions will also be sequentially executed.
func (suite *KeeperTestSuite) TestVerifyFeeAndDeductTxCostsFromUserBalance() {
	var initialBaseFee *big.Int
	hundredInt := sdkmath.NewInt(100)
	oneInt := sdkmath.OneInt()
	fiveInt := sdkmath.NewInt(5)
	fiftyInt := sdkmath.NewInt(50)

	// initial balance enough to cover all test cases that have feemarket enabled
	initBalance := func() sdkmath.Int { return (sdkmath.NewIntFromBigInt(initialBaseFee).AddRaw(10)).MulRaw(105) }

	testSetup := []struct {
		decimals       uint32
		initialBaseFee sdkmath.LegacyDec
	}{
		{evmtypes.Denom18Dec, sdkmath.LegacyNewDec(params.InitialBaseFee)},
		{evmtypes.Denom6Dec, sdkmath.LegacyNewDec(1e13)},
	}
	testCases := []struct {
		name             string
		gasLimit         uint64
		gasPrice         *sdkmath.Int
		gasFeeCap        func() *big.Int
		gasTipCap        *big.Int
		initialBalance   func() sdkmath.Int // initial balance when feemarket is enabled
		cost             *sdkmath.Int
		accessList       *ethtypes.AccessList
		expectPassVerify bool
		expectPassDeduct bool
		enableFeemarket  bool
		from             string
		malleate         func()
	}{
		{
			name:             "Enough balance",
			gasLimit:         10,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             suite.address.String(),
		},
		{
			name:             "Equal balance",
			gasLimit:         100,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             suite.address.String(),
		},
		{
			name:             "Higher gas limit, not enough balance",
			gasLimit:         105,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: false,
			from:             suite.address.String(),
		},
		{
			name:             "Higher gas price, enough balance",
			gasLimit:         20,
			gasPrice:         &fiveInt,
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             suite.address.String(),
		},
		{
			name:             "Higher gas price, not enough balance",
			gasLimit:         20,
			gasPrice:         &fiftyInt,
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: false,
			from:             suite.address.String(),
		},
		// This case is expected to be true because the fees can be deducted, but the tx
		// execution is going to fail because there is no more balance to pay the cost
		{
			name:             "Higher cost, enough balance",
			gasLimit:         100,
			gasPrice:         &oneInt,
			cost:             &fiftyInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             suite.address.String(),
		},
		//  testcases with enableFeemarket enabled.
		{
			name:             "Invalid gasFeeCap w/ enableFeemarket",
			gasLimit:         10,
			gasFeeCap:        func() *big.Int { return big.NewInt(1) },
			gasTipCap:        big.NewInt(1),
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: false,
			expectPassDeduct: false,
			enableFeemarket:  true,
			from:             suite.address.String(),
		},
		{
			name:     "empty tip fee is valid to deduct",
			gasLimit: 10,
			gasFeeCap: func() *big.Int {
				return initialBaseFee
			},
			gasTipCap:        big.NewInt(1),
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             suite.address.String(),
		},
		{
			name:     "effectiveTip equal to gasTipCap",
			gasLimit: 100,
			gasFeeCap: func() *big.Int {
				return initialBaseFee.Add(initialBaseFee, big.NewInt(2))
			},
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             suite.address.String(),
		},
		{
			name:     "effectiveTip equal to (gasFeeCap - baseFee)",
			gasLimit: 105,
			gasFeeCap: func() *big.Int {
				return initialBaseFee.Add(initialBaseFee, big.NewInt(1))
			},
			gasTipCap:        big.NewInt(2),
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             suite.address.String(),
		},
		{
			name:             "Invalid from address",
			gasLimit:         10,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: false,
			from:             "abcdef",
		},
		{
			name:           "Enough balance - with access list",
			gasLimit:       10,
			gasPrice:       &oneInt,
			cost:           &oneInt,
			initialBalance: initBalance,
			accessList: &ethtypes.AccessList{
				ethtypes.AccessTuple{
					Address:     suite.address,
					StorageKeys: []common.Hash{},
				},
			},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             suite.address.String(),
		},
		{
			name:             "gasLimit < intrinsicGas during IsCheckTx",
			gasLimit:         1,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			initialBalance:   initBalance,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: false,
			expectPassDeduct: false,
			from:             suite.address.String(),
			malleate: func() {
				suite.ctx = suite.ctx.WithIsCheckTx(true)
			},
		},
	}

	for _, setup := range testSetup {
		for i, tc := range testCases {
			suite.Run(fmt.Sprintf("%d dec - %s", setup.decimals, tc.name), func() {
				suite.enableFeemarket = tc.enableFeemarket
				suite.denomDecimals = setup.decimals
				suite.SetupTest()
				// update feemarket params (base fee)
				initialBaseFee = setup.initialBaseFee.BigInt()
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.BaseFee = setup.initialBaseFee
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				vmdb := suite.StateDB()

				if tc.malleate != nil {
					tc.malleate()
				}
				var amount, gasPrice, gasFeeCap, gasTipCap, initBal *big.Int

				if tc.cost != nil {
					amount = tc.cost.BigInt()
				}

				if suite.enableFeemarket {
					if tc.gasFeeCap != nil {
						gasFeeCap = tc.gasFeeCap()
					}
					if tc.gasTipCap == nil {
						gasTipCap = oneInt.BigInt()
					} else {
						gasTipCap = tc.gasTipCap
					}
					initBal = tc.initialBalance().BigInt()

				} else {
					if tc.gasPrice != nil {
						gasPrice = tc.gasPrice.BigInt()
					}
					initBal = hundredInt.BigInt()
				}

				if s.denomDecimals == evmtypes.Denom6Dec {
					initBal = evmtypes.Convert6To18DecimalsBigInt(initBal)
					if amount != nil {
						amount = evmtypes.Convert6To18DecimalsBigInt(amount)
					}
					if gasPrice != nil {
						gasPrice = evmtypes.Convert6To18DecimalsBigInt(gasPrice)
					}
					if gasFeeCap != nil {
						gasFeeCap = evmtypes.Convert6To18DecimalsBigInt(gasFeeCap)
					}
					if gasTipCap != nil {
						gasTipCap = evmtypes.Convert6To18DecimalsBigInt(gasTipCap)
					}
				}

				vmdb.AddBalance(suite.address, initBal)
				balance := vmdb.GetBalance(suite.address)
				suite.Require().Equal(balance, initBal)

				err = vmdb.Commit()
				suite.Require().NoError(err, "Unexpected error while committing to vmdb: %d", err)

				ethTxParams := &evmtypes.EvmTxArgs{
					ChainID:   big.NewInt(9000),
					Nonce:     1,
					To:        &suite.address,
					Amount:    amount,
					GasLimit:  tc.gasLimit,
					GasPrice:  gasPrice,
					GasFeeCap: gasFeeCap,
					GasTipCap: gasTipCap,
					Accesses:  tc.accessList,
				}
				tx := evmtypes.NewTx(ethTxParams)
				tx.From = tc.from

				txData, _ := evmtypes.UnpackTxData(tx.Data)

				evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
				baseFee := suite.app.EvmKeeper.GetBaseFee(suite.ctx, evmParams)
				priority := evmtypes.GetTxPriority(txData, baseFee)

				fees, err := keeper.VerifyFee(txData, evmtypes.DefaultEVMDenom, baseFee, false, false, suite.ctx.IsCheckTx())
				if tc.expectPassVerify {
					suite.Require().NoError(err, "valid test %d failed - '%s'", i, tc.name)
					if tc.enableFeemarket {
						baseFeeRes, err := suite.app.EvmKeeper.BaseFee(suite.ctx, &evmtypes.QueryBaseFeeRequest{})
						suite.Require().NoError(err)
						suite.Require().Equal(
							fees,
							sdk.NewCoins(
								sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewIntFromBigInt(txData.EffectiveFee(baseFeeRes.BaseFee.BigInt()))),
							),
							"valid test %d failed, fee value is wrong  - '%s'", i, tc.name,
						)
						if setup.decimals == evmtypes.Denom6Dec {
							// priority will be > 0 due to scaling up the gasTip
							// and feeCap values and keeping the same DefaultPriorityReduction. Considering priority has no effect ATM
							suite.Require().Greater(priority, int64(0))
						} else {
							suite.Require().Equal(int64(0), priority)
						}
					} else {
						suite.Require().Equal(
							fees,
							sdk.NewCoins(
								sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewIntFromBigInt(gasPrice).Mul(sdkmath.NewIntFromUint64(tc.gasLimit))),
							),
							"valid test %d failed, fee value is wrong  - '%s'", i, tc.name,
						)
					}
				} else {
					suite.Require().Error(err, "invalid test %d passed - '%s'", i, tc.name)
					suite.Require().Nil(fees, "invalid test %d passed. fees value must be nil - '%s'", i, tc.name)
				}

				err = suite.app.EvmKeeper.DeductTxCostsFromUserBalance(suite.ctx, fees, common.HexToAddress(tx.From))
				if tc.expectPassDeduct {
					suite.Require().NoError(err, "valid test %d failed - '%s'", i, tc.name)
				} else {
					suite.Require().Error(err, "invalid test %d passed - '%s'", i, tc.name)
				}
			})
		}
	}
	suite.enableFeemarket = false // reset flag
}
