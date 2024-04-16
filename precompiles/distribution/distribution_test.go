package distribution_test

import (
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v18/app"
	"github.com/evmos/evmos/v18/precompiles/distribution"
	"github.com/evmos/evmos/v18/utils"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

func (s *PrecompileTestSuite) TestIsTransaction() {
	testCases := []struct {
		name   string
		method string
		isTx   bool
	}{
		{
			distribution.SetWithdrawAddressMethod,
			s.precompile.Methods[distribution.SetWithdrawAddressMethod].Name,
			true,
		},
		{
			distribution.WithdrawDelegatorRewardsMethod,
			s.precompile.Methods[distribution.WithdrawDelegatorRewardsMethod].Name,
			true,
		},
		{
			distribution.WithdrawValidatorCommissionMethod,
			s.precompile.Methods[distribution.WithdrawValidatorCommissionMethod].Name,
			true,
		},
		{
			distribution.ValidatorDistributionInfoMethod,
			s.precompile.Methods[distribution.ValidatorDistributionInfoMethod].Name,
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

// TestRun tests the precompile's Run method.
func (s *PrecompileTestSuite) TestRun() {
	testcases := []struct {
		name        string
		malleate    func() (common.Address, []byte)
		readOnly    bool
		expPass     bool
		errContains string
	}{
		{
			name: "pass - set withdraw address transaction",
			malleate: func() (common.Address, []byte) {
				valAddr, err := sdk.ValAddressFromBech32(s.validators[0].OperatorAddress)
				s.Require().NoError(err)
				val, _ := s.app.StakingKeeper.GetValidator(s.ctx, valAddr)
				coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
				s.app.DistrKeeper.AllocateTokensToValidator(s.ctx, val, sdk.NewDecCoinsFromCoins(coins...))

				input, err := s.precompile.Pack(
					distribution.SetWithdrawAddressMethod,
					s.address,
					s.address.String(),
				)
				s.Require().NoError(err, "failed to pack input")
				return s.address, input
			},
			readOnly: false,
			expPass:  true,
		},
		{
			name: "pass - withdraw validator commissions transaction",
			malleate: func() (common.Address, []byte) {
				hexAddr := common.Bytes2Hex(s.address.Bytes())
				valAddr, err := sdk.ValAddressFromHex(hexAddr)
				s.Require().NoError(err)
				caller := common.BytesToAddress(valAddr)

				valCommission := sdk.DecCoins{sdk.NewDecCoinFromDec(utils.BaseDenom, math.LegacyNewDecWithPrec(1000000000000000000, 1))}
				// set outstanding rewards
				s.app.DistrKeeper.SetValidatorOutstandingRewards(s.ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: valCommission})
				// set commission
				s.app.DistrKeeper.SetValidatorAccumulatedCommission(s.ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: valCommission})

				input, err := s.precompile.Pack(
					distribution.WithdrawValidatorCommissionMethod,
					valAddr.String(),
				)
				s.Require().NoError(err, "failed to pack input")
				return caller, input
			},
			readOnly: false,
			expPass:  true,
		},
		{
			name: "pass - withdraw delegator rewards transaction",
			malleate: func() (common.Address, []byte) {
				valAddr, err := sdk.ValAddressFromBech32(s.validators[0].OperatorAddress)
				s.Require().NoError(err)
				val, _ := s.app.StakingKeeper.GetValidator(s.ctx, valAddr)
				coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
				s.app.DistrKeeper.AllocateTokensToValidator(s.ctx, val, sdk.NewDecCoinsFromCoins(coins...))

				input, err := s.precompile.Pack(
					distribution.WithdrawDelegatorRewardsMethod,
					s.address,
					valAddr.String(),
				)
				s.Require().NoError(err, "failed to pack input")

				return s.address, input
			},
			readOnly: false,
			expPass:  true,
		},
		{
			name: "pass - claim rewards transaction",
			malleate: func() (common.Address, []byte) {
				valAddr, err := sdk.ValAddressFromBech32(s.validators[0].OperatorAddress)
				s.Require().NoError(err)
				val, _ := s.app.StakingKeeper.GetValidator(s.ctx, valAddr)
				coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
				s.app.DistrKeeper.AllocateTokensToValidator(s.ctx, val, sdk.NewDecCoinsFromCoins(coins...))

				input, err := s.precompile.Pack(
					distribution.ClaimRewardsMethod,
					s.address,
					uint32(2),
				)
				s.Require().NoError(err, "failed to pack input")

				return s.address, input
			},
			readOnly: false,
			expPass:  true,
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			// setup basic test suite
			s.SetupTest()

			baseFee := s.app.FeeMarketKeeper.GetBaseFee(s.ctx)

			// malleate testcase
			caller, input := tc.malleate()

			contract := vm.NewPrecompile(vm.AccountRef(caller), s.precompile, big.NewInt(0), uint64(1e6))
			contract.Input = input

			contractAddr := contract.Address()
			// Build and sign Ethereum transaction
			txArgs := evmtypes.EvmTxArgs{
				ChainID:   s.app.EvmKeeper.ChainID(),
				Nonce:     0,
				To:        &contractAddr,
				Amount:    nil,
				GasLimit:  100000,
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
			}
		})
	}
}
