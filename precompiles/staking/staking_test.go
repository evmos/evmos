package staking_test

import (
	"math/big"
	"time"

	"github.com/evmos/evmos/v13/app"

	"github.com/evmos/evmos/v13/precompiles/authorization"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v13/precompiles/staking"
	"github.com/evmos/evmos/v13/utils"
	evmtypes "github.com/evmos/evmos/v13/x/evm/types"
)

func (s *PrecompileTestSuite) TestIsTransaction() {
	testCases := []struct {
		name   string
		method string
		isTx   bool
	}{
		{
			authorization.ApproveMethod,
			s.precompile.Methods[authorization.ApproveMethod].Name,
			true,
		},
		{
			authorization.IncreaseAllowanceMethod,
			s.precompile.Methods[authorization.IncreaseAllowanceMethod].Name,
			true,
		},
		{
			authorization.DecreaseAllowanceMethod,
			s.precompile.Methods[authorization.DecreaseAllowanceMethod].Name,
			true,
		},
		{
			staking.DelegateMethod,
			s.precompile.Methods[staking.DelegateMethod].Name,
			true,
		},
		{
			staking.UndelegateMethod,
			s.precompile.Methods[staking.UndelegateMethod].Name,
			true,
		},
		{
			staking.RedelegateMethod,
			s.precompile.Methods[staking.RedelegateMethod].Name,
			true,
		},
		{
			staking.CancelUnbondingDelegationMethod,
			s.precompile.Methods[staking.CancelUnbondingDelegationMethod].Name,
			true,
		},
		{
			staking.DelegationMethod,
			s.precompile.Methods[staking.DelegationMethod].Name,
			false,
		},
		{
			"invalid",
			"invalid",
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(s.precompile.IsTransaction(tc.method), tc.isTx)
		})
	}
}

func (s *PrecompileTestSuite) TestRequiredGas() {
	testcases := []struct {
		name     string
		malleate func() []byte
		expGas   uint64
	}{
		{
			"success - delegate transaction with correct gas estimation",
			func() []byte {
				input, err := s.precompile.Pack(
					staking.DelegateMethod,
					s.address,
					s.validators[0].GetOperator().String(),
					big.NewInt(10000000000),
				)
				s.Require().NoError(err)
				return input
			},
			7760,
		},
		{
			"success - undelegate transaction with correct gas estimation",
			func() []byte {
				input, err := s.precompile.Pack(
					staking.UndelegateMethod,
					s.address,
					s.validators[0].GetOperator().String(),
					big.NewInt(1),
				)
				s.Require().NoError(err)
				return input
			},
			7760,
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// malleate contract input
			input := tc.malleate()
			gas := s.precompile.RequiredGas(input)

			s.Require().Equal(gas, tc.expGas)
		})
	}
}

// TestRun tests the precompile's Run method.
func (s *PrecompileTestSuite) TestRun() {
	testcases := []struct {
		name        string
		malleate    func() []byte
		gas         uint64
		readOnly    bool
		expPass     bool
		errContains string
	}{
		{
			"fail - contract gas limit is < gas cost to run a query / tx",
			func() []byte {
				err := s.CreateAuthorization(s.address, staking.DelegateAuthz, nil)
				s.Require().NoError(err)

				input, err := s.precompile.Pack(
					staking.DelegateMethod,
					s.address,
					s.validators[0].GetOperator().String(),
					big.NewInt(1000),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			8000,
			false,
			false,
			"out of gas",
		},
		{
			"pass - delegate transaction",
			func() []byte {
				err := s.CreateAuthorization(s.address, staking.DelegateAuthz, nil)
				s.Require().NoError(err)

				input, err := s.precompile.Pack(
					staking.DelegateMethod,
					s.address,
					s.validators[0].GetOperator().String(),
					big.NewInt(1000),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			1000000,
			false,
			true,
			"",
		},
		{
			"pass - undelegate transaction",
			func() []byte {
				err := s.CreateAuthorization(s.address, staking.UndelegateAuthz, nil)
				s.Require().NoError(err)

				input, err := s.precompile.Pack(
					staking.UndelegateMethod,
					s.address,
					s.validators[0].GetOperator().String(),
					big.NewInt(1),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			1000000,
			false,
			true,
			"",
		},
		{
			"pass - redelegate transaction",
			func() []byte {
				err := s.CreateAuthorization(s.address, staking.RedelegateAuthz, nil)
				s.Require().NoError(err)

				input, err := s.precompile.Pack(
					staking.RedelegateMethod,
					s.address,
					s.validators[0].GetOperator().String(),
					s.validators[1].GetOperator().String(),
					big.NewInt(1),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			1000000,
			false,
			true,
			"failed to redelegate tokens",
		},
		{
			"pass - cancel unbonding delegation transaction",
			func() []byte {
				// add unbonding delegation to staking keeper
				ubd := stakingtypes.NewUnbondingDelegation(
					s.address.Bytes(),
					s.validators[0].GetOperator(),
					1000,
					time.Now().Add(time.Hour),
					sdk.NewInt(1000),
				)
				s.app.StakingKeeper.SetUnbondingDelegation(s.ctx, ubd)

				err := s.CreateAuthorization(s.address, staking.CancelUnbondingDelegationAuthz, nil)
				s.Require().NoError(err)

				// Needs to be called after setting unbonding delegation
				// In order to mimic the coins being added to the unboding pool
				coin := sdk.NewCoin(utils.BaseDenom, sdk.NewInt(1000))
				err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, sdk.Coins{coin})
				s.Require().NoError(err, "failed to send coins from module to module")

				input, err := s.precompile.Pack(
					staking.CancelUnbondingDelegationMethod,
					s.address,
					s.validators[0].GetOperator().String(),
					big.NewInt(1000),
					big.NewInt(1000),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			1000000,
			false,
			true,
			"",
		},
		{
			"pass - delegation query",
			func() []byte {
				input, err := s.precompile.Pack(
					staking.DelegationMethod,
					s.address,
					s.validators[0].GetOperator().String(),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			1000000,
			false,
			true,
			"",
		},
		{
			"pass - validator query",
			func() []byte {
				input, err := s.precompile.Pack(
					staking.ValidatorMethod,
					s.validators[0].OperatorAddress,
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			1000000,
			false,
			true,
			"",
		},
		{
			"pass - redelgation query",
			func() []byte {
				// add redelegation to staking keeper
				redelegation := stakingtypes.NewRedelegation(
					s.address.Bytes(),
					s.validators[0].GetOperator(),
					s.validators[1].GetOperator(),
					1000,
					time.Now().Add(time.Hour),
					sdk.NewInt(1000),
					sdk.NewDec(1),
				)

				s.app.StakingKeeper.SetRedelegation(s.ctx, redelegation)

				input, err := s.precompile.Pack(
					staking.RedelegationMethod,
					s.address,
					s.validators[0].GetOperator().String(),
					s.validators[1].GetOperator().String(),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			1000000,
			false,
			true,
			"",
		},
		{
			"pass - delegation query - read only",
			func() []byte {
				input, err := s.precompile.Pack(
					staking.DelegationMethod,
					s.address,
					s.validators[0].GetOperator().String(),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			1000000,
			true,
			true,
			"",
		},
		{
			"pass - unbonding delegation query",
			func() []byte {
				// add unbonding delegation to staking keeper
				ubd := stakingtypes.NewUnbondingDelegation(
					s.address.Bytes(),
					s.validators[0].GetOperator(),
					1000,
					time.Now().Add(time.Hour),
					sdk.NewInt(1000),
				)
				s.app.StakingKeeper.SetUnbondingDelegation(s.ctx, ubd)

				// Needs to be called after setting unbonding delegation
				// In order to mimic the coins being added to the unboding pool
				coin := sdk.NewCoin(utils.BaseDenom, sdk.NewInt(1000))
				err := s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, sdk.Coins{coin})
				s.Require().NoError(err, "failed to send coins from module to module")

				input, err := s.precompile.Pack(
					staking.UnbondingDelegationMethod,
					s.address,
					s.validators[0].GetOperator().String(),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			1000000,
			true,
			true,
			"",
		},
		{
			"fail - delegate method - read only",
			func() []byte {
				input, err := s.precompile.Pack(
					staking.DelegateMethod,
					s.address,
					s.validators[0].GetOperator().String(),
					big.NewInt(1000),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			0,
			true,
			false,
			"write protection",
		},
		{
			"fail - invalid method",
			func() []byte {
				return []byte("invalid")
			},
			0,
			false,
			false,
			"no method with id",
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			// setup basic test suite
			s.SetupTest()

			baseFee := s.app.FeeMarketKeeper.GetBaseFee(s.ctx)

			contract := vm.NewPrecompile(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)
			contractAddr := contract.Address()

			// malleate testcase
			contract.Input = tc.malleate()

			// Build and sign Ethereum transaction
			txArgs := evmtypes.EvmTxArgs{
				ChainID:   s.app.EvmKeeper.ChainID(),
				Nonce:     0,
				To:        &contractAddr,
				Amount:    nil,
				GasLimit:  tc.gas,
				GasPrice:  app.MainnetMinGasPrices.BigInt(),
				GasFeeCap: baseFee,
				GasTipCap: big.NewInt(1),
				Accesses:  &ethtypes.AccessList{},
			}
			msgEthereumTx := evmtypes.NewTx(&txArgs)

			msgEthereumTx.From = s.address.String()
			err := msgEthereumTx.Sign(s.ethSigner, s.signer)
			s.Require().NoError(err, "failed to sign Ethereum message")

			// Instantiate config
			proposerAddress := s.ctx.BlockHeader().ProposerAddress
			cfg, err := s.app.EvmKeeper.EVMConfig(s.ctx, proposerAddress, s.app.EvmKeeper.ChainID())
			s.Require().NoError(err, "failed to instantiate EVM config")

			msg, err := msgEthereumTx.AsMessage(s.ethSigner, baseFee)
			s.Require().NoError(err, "failed to instantiate Ethereum message")

			// Instantiate EVM
			evm := s.app.EvmKeeper.NewEVM(
				s.ctx, msg, cfg, nil, s.stateDB,
			)

			params := s.app.EvmKeeper.GetParams(s.ctx)
			activePrecompiles := params.GetActivePrecompilesAddrs()
			precompileMap := s.app.EvmKeeper.Precompiles(activePrecompiles...)
			err = vm.ValidatePrecompiles(precompileMap, activePrecompiles)
			s.Require().NoError(err, "invalid precompiles", activePrecompiles)
			evm.WithPrecompiles(precompileMap, activePrecompiles)

			// Run precompiled contract
			bz, err := s.precompile.Run(evm, contract, tc.readOnly)

			// Check results
			if tc.expPass {
				s.Require().NoError(err, "expected no error when running the precompile")
				s.Require().NotNil(bz, "expected returned bytes not to be nil")
			} else {
				s.Require().Error(err, "expected error to be returned when running the precompile")
				s.Require().Nil(bz, "expected returned bytes to be nil")
				s.Require().ErrorContains(err, tc.errContains)
				consumed := s.ctx.GasMeter().GasConsumed()
				// LessThanOrEqual because the gas is consumed before the error is returned
				s.Require().LessOrEqual(tc.gas, consumed, "expected gas consumed to be equal to gas limit")

			}
		})
	}
}
