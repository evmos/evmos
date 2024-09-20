package keeper_test

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethparams "github.com/ethereum/go-ethereum/params"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/x/evm/config"
	"github.com/evmos/evmos/v20/x/evm/keeper"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

func (suite *KeeperTestSuite) TestCheckSenderBalance() {
	hundredInt := sdkmath.NewInt(100)
	zeroInt := sdkmath.ZeroInt()
	oneInt := sdkmath.OneInt()
	fiveInt := sdkmath.NewInt(5)
	fiftyInt := sdkmath.NewInt(50)
	negInt := sdkmath.NewInt(-10)
	addr := utiltx.GenerateAddress()

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
			to:         suite.keyring.GetAddr(0).String(),
			gasLimit:   10,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       addr.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Equal balance",
			to:         suite.keyring.GetAddr(0).String(),
			gasLimit:   99,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       addr.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "negative cost",
			to:         suite.keyring.GetAddr(0).String(),
			gasLimit:   1,
			gasPrice:   &oneInt,
			cost:       &negInt,
			from:       addr.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher gas limit, not enough balance",
			to:         suite.keyring.GetAddr(0).String(),
			gasLimit:   100,
			gasPrice:   &oneInt,
			cost:       &oneInt,
			from:       addr.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher gas price, enough balance",
			to:         suite.keyring.GetAddr(0).String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &oneInt,
			from:       addr.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Higher gas price, not enough balance",
			to:         suite.keyring.GetAddr(0).String(),
			gasLimit:   20,
			gasPrice:   &fiveInt,
			cost:       &oneInt,
			from:       addr.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:       "Higher cost, enough balance",
			to:         suite.keyring.GetAddr(0).String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &fiftyInt,
			from:       addr.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: true,
		},
		{
			name:       "Higher cost, not enough balance",
			to:         suite.keyring.GetAddr(0).String(),
			gasLimit:   10,
			gasPrice:   &fiveInt,
			cost:       &hundredInt,
			from:       addr.String(),
			accessList: &ethtypes.AccessList{},
			expectPass: false,
		},
		{
			name:            "Enough balance w/ enableFeemarket",
			to:              suite.keyring.GetAddr(0).String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(1),
			cost:            &oneInt,
			from:            addr.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "Equal balance w/ enableFeemarket",
			to:              suite.keyring.GetAddr(0).String(),
			gasLimit:        99,
			gasFeeCap:       big.NewInt(1),
			cost:            &oneInt,
			from:            addr.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "negative cost w/ enableFeemarket",
			to:              suite.keyring.GetAddr(0).String(),
			gasLimit:        1,
			gasFeeCap:       big.NewInt(1),
			cost:            &negInt,
			from:            addr.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
		{
			name:            "Higher gas limit, not enough balance w/ enableFeemarket",
			to:              suite.keyring.GetAddr(0).String(),
			gasLimit:        100,
			gasFeeCap:       big.NewInt(1),
			cost:            &oneInt,
			from:            addr.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
		{
			name:            "Higher gas price, enough balance w/ enableFeemarket",
			to:              suite.keyring.GetAddr(0).String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(5),
			cost:            &oneInt,
			from:            addr.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "Higher gas price, not enough balance w/ enableFeemarket",
			to:              suite.keyring.GetAddr(0).String(),
			gasLimit:        20,
			gasFeeCap:       big.NewInt(5),
			cost:            &oneInt,
			from:            addr.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
		{
			name:            "Higher cost, enough balance w/ enableFeemarket",
			to:              suite.keyring.GetAddr(0).String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(5),
			cost:            &fiftyInt,
			from:            addr.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      true,
			enableFeemarket: true,
		},
		{
			name:            "Higher cost, not enough balance w/ enableFeemarket",
			to:              suite.keyring.GetAddr(0).String(),
			gasLimit:        10,
			gasFeeCap:       big.NewInt(5),
			cost:            &hundredInt,
			from:            addr.String(),
			accessList:      &ethtypes.AccessList{},
			expectPass:      false,
			enableFeemarket: true,
		},
	}

	vmdb := suite.StateDB()
	vmdb.AddBalance(addr, hundredInt.BigInt())
	balance := vmdb.GetBalance(addr)
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

			acct := suite.network.App.EvmKeeper.GetAccountOrEmpty(suite.network.GetContext(), addr)
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
	hundredInt := sdkmath.NewInt(100)
	zeroInt := sdkmath.ZeroInt()
	oneInt := sdkmath.NewInt(1)
	fiveInt := sdkmath.NewInt(5)
	fiftyInt := sdkmath.NewInt(50)
	addr, _ := utiltx.NewAddrKey()

	// should be enough to cover all test cases
	initBalance := sdkmath.NewInt((ethparams.InitialBaseFee + 10) * 105)

	testCases := []struct {
		name             string
		gasLimit         uint64
		gasPrice         *sdkmath.Int
		gasFeeCap        *big.Int
		gasTipCap        *big.Int
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
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             addr.String(),
		},
		{
			name:             "Equal balance",
			gasLimit:         100,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             addr.String(),
		},
		{
			name:             "Higher gas limit, not enough balance",
			gasLimit:         105,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: false,
			from:             addr.String(),
		},
		{
			name:             "Higher gas price, enough balance",
			gasLimit:         20,
			gasPrice:         &fiveInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             addr.String(),
		},
		{
			name:             "Higher gas price, not enough balance",
			gasLimit:         20,
			gasPrice:         &fiftyInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: false,
			from:             addr.String(),
		},
		// This case is expected to be true because the fees can be deducted, but the tx
		// execution is going to fail because there is no more balance to pay the cost
		{
			name:             "Higher cost, enough balance",
			gasLimit:         100,
			gasPrice:         &oneInt,
			cost:             &fiftyInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             addr.String(),
		},
		//  testcases with enableFeemarket enabled.
		{
			name:             "Invalid gasFeeCap w/ enableFeemarket",
			gasLimit:         10,
			gasFeeCap:        big.NewInt(1),
			gasTipCap:        big.NewInt(1),
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: false,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             addr.String(),
		},
		{
			name:             "empty tip fee is valid to deduct",
			gasLimit:         10,
			gasFeeCap:        big.NewInt(ethparams.InitialBaseFee),
			gasTipCap:        big.NewInt(1),
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             addr.String(),
		},
		{
			name:             "effectiveTip equal to gasTipCap",
			gasLimit:         100,
			gasFeeCap:        big.NewInt(ethparams.InitialBaseFee + 2),
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             addr.String(),
		},
		{
			name:             "effectiveTip equal to (gasFeeCap - baseFee)",
			gasLimit:         105,
			gasFeeCap:        big.NewInt(ethparams.InitialBaseFee + 1),
			gasTipCap:        big.NewInt(2),
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: true,
			enableFeemarket:  true,
			from:             addr.String(),
		},
		{
			name:             "Invalid from address",
			gasLimit:         10,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: true,
			expectPassDeduct: false,
			from:             "abcdef",
		},
		{
			name:     "Enough balance - with access list",
			gasLimit: 10,
			gasPrice: &oneInt,
			cost:     &oneInt,
			accessList: &ethtypes.AccessList{
				ethtypes.AccessTuple{
					Address:     suite.keyring.GetAddr(0),
					StorageKeys: []common.Hash{},
				},
			},
			expectPassVerify: true,
			expectPassDeduct: true,
			from:             addr.String(),
		},
		{
			name:             "gasLimit < intrinsicGas during IsCheckTx",
			gasLimit:         1,
			gasPrice:         &oneInt,
			cost:             &oneInt,
			accessList:       &ethtypes.AccessList{},
			expectPassVerify: false,
			expectPassDeduct: true,
			from:             addr.String(),
			malleate: func() {
				suite.network.WithIsCheckTxCtx(true)
			},
		},
	}

	for i, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.enableFeemarket = tc.enableFeemarket
			suite.SetupTest()
			vmdb := suite.StateDB()

			if tc.malleate != nil {
				tc.malleate()
			}
			var amount, gasPrice, gasFeeCap, gasTipCap *big.Int
			if tc.cost != nil {
				amount = tc.cost.BigInt()
			}

			if suite.enableFeemarket {
				if tc.gasFeeCap != nil {
					gasFeeCap = tc.gasFeeCap
				}
				if tc.gasTipCap == nil {
					gasTipCap = oneInt.BigInt()
				} else {
					gasTipCap = tc.gasTipCap
				}
				vmdb.AddBalance(addr, initBalance.BigInt())
				balance := vmdb.GetBalance(addr)
				suite.Require().Equal(balance, initBalance.BigInt())
			} else {
				if tc.gasPrice != nil {
					gasPrice = tc.gasPrice.BigInt()
				}

				vmdb.AddBalance(addr, hundredInt.BigInt())
				balance := vmdb.GetBalance(addr)
				suite.Require().Equal(balance, hundredInt.BigInt())
			}
			err := vmdb.Commit()
			suite.Require().NoError(err, "Unexpected error while committing to vmdb: %d", err)

			toAddr := suite.keyring.GetAddr(0)
			ethTxParams := &evmtypes.EvmTxArgs{
				ChainID:   zeroInt.BigInt(),
				Nonce:     1,
				To:        &toAddr,
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

			ethCfg := config.GetChainConfig()
			baseFee := suite.network.App.EvmKeeper.GetBaseFee(suite.network.GetContext(), ethCfg)
			priority := evmtypes.GetTxPriority(txData, baseFee)

			baseDenom := config.GetDenom()

			fees, err := keeper.VerifyFee(txData, baseDenom, baseFee, false, false, suite.network.GetContext().IsCheckTx())
			if tc.expectPassVerify {
				suite.Require().NoError(err, "valid test %d failed - '%s'", i, tc.name)
				if tc.enableFeemarket {
					baseFee := suite.network.App.FeeMarketKeeper.GetBaseFee(suite.network.GetContext())
					suite.Require().Equal(
						fees,
						sdk.NewCoins(
							sdk.NewCoin(baseDenom, sdkmath.NewIntFromBigInt(txData.EffectiveFee(baseFee))),
						),
						"valid test %d failed, fee value is wrong  - '%s'", i, tc.name,
					)
					suite.Require().Equal(int64(0), priority)
				} else {
					suite.Require().Equal(
						fees,
						sdk.NewCoins(
							sdk.NewCoin(baseDenom, tc.gasPrice.Mul(sdkmath.NewIntFromUint64(tc.gasLimit))),
						),
						"valid test %d failed, fee value is wrong  - '%s'", i, tc.name,
					)
				}
			} else {
				suite.Require().Error(err, "invalid test %d passed - '%s'", i, tc.name)
				suite.Require().Nil(fees, "invalid test %d passed. fees value must be nil - '%s'", i, tc.name)
			}

			err = suite.network.App.EvmKeeper.DeductTxCostsFromUserBalance(suite.network.GetContext(), fees, common.HexToAddress(tx.From))
			if tc.expectPassDeduct {
				suite.Require().NoError(err, "valid test %d failed - '%s'", i, tc.name)
			} else {
				suite.Require().Error(err, "invalid test %d passed - '%s'", i, tc.name)
			}
		})
	}
	suite.enableFeemarket = false // reset flag
}
