package staking_test

import (
	"math/big"
	"time"

	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"

	"cosmossdk.io/math"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v19/x/evm/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v19/app"
	"github.com/evmos/evmos/v19/precompiles/authorization"
	"github.com/evmos/evmos/v19/precompiles/staking"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
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
			staking.CreateValidatorMethod,
			s.precompile.Methods[staking.CreateValidatorMethod].Name,
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
					s.keyring.GetAddr(0),
					s.network.GetValidators()[0].GetOperator(),
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
					s.keyring.GetAddr(0),
					s.network.GetValidators()[0].GetOperator(),
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
	var ctx sdk.Context
	testcases := []struct {
		name        string
		malleate    func(delegator, grantee testkeyring.Key) []byte
		gas         uint64
		readOnly    bool
		expPass     bool
		errContains string
	}{
		{
			"fail - contract gas limit is < gas cost to run a query / tx",
			func(delegator, grantee testkeyring.Key) []byte {
				// TODO: why is this required?
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.DelegateAuthz, nil)
				s.Require().NoError(err)

				input, err := s.precompile.Pack(
					staking.DelegateMethod,
					delegator.Addr,
					s.network.GetValidators()[0].GetOperator(),
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
			func(delegator, grantee testkeyring.Key) []byte {
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.DelegateAuthz, nil)
				s.Require().NoError(err)

				input, err := s.precompile.Pack(
					staking.DelegateMethod,
					delegator.Addr,
					s.network.GetValidators()[0].GetOperator(),
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
			func(delegator, grantee testkeyring.Key) []byte {
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.UndelegateAuthz, nil)
				s.Require().NoError(err)

				input, err := s.precompile.Pack(
					staking.UndelegateMethod,
					delegator.Addr,
					s.network.GetValidators()[0].GetOperator(),
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
			func(delegator, grantee testkeyring.Key) []byte {
				err := s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.RedelegateAuthz, nil)
				s.Require().NoError(err)

				input, err := s.precompile.Pack(
					staking.RedelegateMethod,
					delegator.Addr,
					s.network.GetValidators()[0].GetOperator(),
					s.network.GetValidators()[1].GetOperator(),
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
			func(delegator, grantee testkeyring.Key) []byte {
				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)
				// add unbonding delegation to staking keeper
				ubd := stakingtypes.NewUnbondingDelegation(
					delegator.AccAddr,
					valAddr,
					ctx.BlockHeight(),
					time.Now().Add(time.Hour),
					math.NewInt(1000),
					0,
					s.network.App.StakingKeeper.ValidatorAddressCodec(),
					s.network.App.AccountKeeper.AddressCodec(),
				)
				err = s.network.App.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
				s.Require().NoError(err, "failed to set unbonding delegation")

				err = s.CreateAuthorization(ctx, delegator.AccAddr, grantee.AccAddr, staking.CancelUnbondingDelegationAuthz, nil)
				s.Require().NoError(err)

				// Needs to be called after setting unbonding delegation
				// In order to mimic the coins being added to the unboding pool
				coin := sdk.NewCoin(utils.BaseDenom, math.NewInt(1000))
				err = s.network.App.BankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, sdk.Coins{coin})
				s.Require().NoError(err, "failed to send coins from module to module")

				input, err := s.precompile.Pack(
					staking.CancelUnbondingDelegationMethod,
					delegator.Addr,
					s.network.GetValidators()[0].GetOperator(),
					big.NewInt(1000),
					big.NewInt(ctx.BlockHeight()),
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
			func(delegator, _ testkeyring.Key) []byte {
				input, err := s.precompile.Pack(
					staking.DelegationMethod,
					delegator.Addr,
					s.network.GetValidators()[0].GetOperator(),
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
			func(_, _ testkeyring.Key) []byte {
				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
				s.Require().NoError(err)

				input, err := s.precompile.Pack(
					staking.ValidatorMethod,
					common.BytesToAddress(valAddr.Bytes()),
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
			func(delegator, _ testkeyring.Key) []byte {
				valAddr1, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)
				valAddr2, err := sdk.ValAddressFromBech32(s.network.GetValidators()[1].GetOperator())
				s.Require().NoError(err)
				// add redelegation to staking keeper
				redelegation := stakingtypes.NewRedelegation(
					delegator.AccAddr,
					valAddr1,
					valAddr2,
					ctx.BlockHeight(),
					time.Now().Add(time.Hour),
					math.NewInt(1000),
					math.LegacyNewDec(1),
					0,
					s.network.App.StakingKeeper.ValidatorAddressCodec(),
					s.network.App.AccountKeeper.AddressCodec(),
				)

				err = s.network.App.StakingKeeper.SetRedelegation(ctx, redelegation)
				s.Require().NoError(err, "failed to set redelegation")

				input, err := s.precompile.Pack(
					staking.RedelegationMethod,
					delegator.Addr,
					s.network.GetValidators()[0].GetOperator(),
					s.network.GetValidators()[1].GetOperator(),
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
			func(delegator, _ testkeyring.Key) []byte {
				input, err := s.precompile.Pack(
					staking.DelegationMethod,
					delegator.Addr,
					s.network.GetValidators()[0].GetOperator(),
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
			func(delegator, _ testkeyring.Key) []byte {
				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)
				// add unbonding delegation to staking keeper
				ubd := stakingtypes.NewUnbondingDelegation(
					delegator.AccAddr,
					valAddr,
					ctx.BlockHeight(),
					time.Now().Add(time.Hour),
					math.NewInt(1000),
					0,
					s.network.App.StakingKeeper.ValidatorAddressCodec(),
					s.network.App.AccountKeeper.AddressCodec(),
				)
				err = s.network.App.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
				s.Require().NoError(err, "failed to set unbonding delegation")

				// Needs to be called after setting unbonding delegation
				// In order to mimic the coins being added to the unboding pool
				coin := sdk.NewCoin(utils.BaseDenom, math.NewInt(1000))
				err = s.network.App.BankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, sdk.Coins{coin})
				s.Require().NoError(err, "failed to send coins from module to module")

				input, err := s.precompile.Pack(
					staking.UnbondingDelegationMethod,
					delegator.Addr,
					s.network.GetValidators()[0].GetOperator(),
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
			func(delegator, _ testkeyring.Key) []byte {
				input, err := s.precompile.Pack(
					staking.DelegateMethod,
					delegator.Addr,
					s.network.GetValidators()[0].GetOperator(),
					big.NewInt(1000),
				)
				s.Require().NoError(err, "failed to pack input")
				return input
			},
			1, // use gas > 0 to avoid doing gas estimation
			true,
			false,
			"write protection",
		},
		{
			"fail - invalid method",
			func(_, _ testkeyring.Key) []byte {
				return []byte("invalid")
			},
			1, // use gas > 0 to avoid doing gas estimation
			false,
			false,
			"no method with id",
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			// setup basic test suite
			s.SetupTest()
			ctx = s.network.GetContext().WithBlockTime(time.Now())

			baseFee := s.network.App.FeeMarketKeeper.GetBaseFee(ctx)

			delegator := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			contract := vm.NewPrecompile(vm.AccountRef(delegator.Addr), s.precompile, big.NewInt(0), tc.gas)
			contractAddr := contract.Address()

			// malleate testcase
			contract.Input = tc.malleate(delegator, grantee)

			// Build and sign Ethereum transaction
			txArgs := evmtypes.EvmTxArgs{
				ChainID:   s.network.App.EvmKeeper.ChainID(),
				Nonce:     0,
				To:        &contractAddr,
				Amount:    nil,
				GasLimit:  tc.gas,
				GasPrice:  app.MainnetMinGasPrices.BigInt(),
				GasFeeCap: baseFee,
				GasTipCap: big.NewInt(1),
				Accesses:  &ethtypes.AccessList{},
			}

			msg, err := s.factory.GenerateGethCoreMsg(delegator.Priv, txArgs)
			s.Require().NoError(err)

			// Instantiate config
			proposerAddress := ctx.BlockHeader().ProposerAddress
			cfg, err := s.network.App.EvmKeeper.EVMConfig(ctx, proposerAddress, s.network.App.EvmKeeper.ChainID())
			s.Require().NoError(err, "failed to instantiate EVM config")

			// Instantiate EVM
			headerHash := ctx.HeaderHash()
			stDB := statedb.New(
				ctx,
				s.network.App.EvmKeeper,
				statedb.NewEmptyTxConfig(common.BytesToHash(headerHash)),
			)
			evm := s.network.App.EvmKeeper.NewEVM(
				ctx, msg, cfg, nil, stDB,
			)

			precompiles, found, err := s.network.App.EvmKeeper.GetPrecompileInstance(ctx, contractAddr)
			s.Require().NoError(err, "failed to instantiate precompile")
			s.Require().True(found, "not found precompile")
			evm.WithPrecompiles(precompiles.Map, precompiles.Addresses)

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
				consumed := ctx.GasMeter().GasConsumed()
				// LessThanOrEqual because the gas is consumed before the error is returned
				s.Require().LessOrEqual(tc.gas, consumed, "expected gas consumed to be equal to gas limit")

			}
		})
	}
}
