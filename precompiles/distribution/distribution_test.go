package distribution_test

import (
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v20/app"
	"github.com/evmos/evmos/v20/precompiles/distribution"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

func (s *PrecompileTestSuite) TestIsTransaction() {
	testCases := []struct {
		name   string
		method abi.Method
		isTx   bool
	}{
		{
			distribution.SetWithdrawAddressMethod,
			s.precompile.Methods[distribution.SetWithdrawAddressMethod],
			true,
		},
		{
			distribution.WithdrawDelegatorRewardsMethod,
			s.precompile.Methods[distribution.WithdrawDelegatorRewardsMethod],
			true,
		},
		{
			distribution.WithdrawValidatorCommissionMethod,
			s.precompile.Methods[distribution.WithdrawValidatorCommissionMethod],
			true,
		},
		{
			distribution.FundCommunityPoolMethod,
			s.precompile.Methods[distribution.FundCommunityPoolMethod],
			true,
		},
		{
			distribution.ValidatorDistributionInfoMethod,
			s.precompile.Methods[distribution.ValidatorDistributionInfoMethod],
			false,
		},
		{
			"invalid",
			abi.Method{},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(s.precompile.IsTransaction(&tc.method), tc.isTx)
		})
	}
}

// TestRun tests the precompile's Run method.
func (s *PrecompileTestSuite) TestRun() {
	var (
		ctx sdk.Context
		err error
	)
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
				valAddr, err := sdk.ValAddressFromBech32(s.network.GetValidators()[0].OperatorAddress)
				s.Require().NoError(err)
				val, _ := s.network.App.StakingKeeper.GetValidator(ctx, valAddr)
				coins := sdk.NewCoins(sdk.NewCoin(s.baseDenom, math.NewInt(1e18)))
				s.Require().NoError(s.network.App.DistrKeeper.AllocateTokensToValidator(ctx, val, sdk.NewDecCoinsFromCoins(coins...)))

				input, err := s.precompile.Pack(
					distribution.SetWithdrawAddressMethod,
					s.keyring.GetAddr(0),
					s.keyring.GetAddr(0).String(),
				)
				s.Require().NoError(err, "failed to pack input")
				return s.keyring.GetAddr(0), input
			},
			readOnly: false,
			expPass:  true,
		},
		{
			name: "pass - withdraw validator commissions transaction",
			malleate: func() (common.Address, []byte) {
				hexAddr := common.Bytes2Hex(s.keyring.GetAddr(0).Bytes())
				valAddr, err := sdk.ValAddressFromHex(hexAddr)
				s.Require().NoError(err)
				caller := common.BytesToAddress(valAddr)

				commAmt := math.LegacyNewDecWithPrec(1000000000000000000, 1)
				valCommission := sdk.DecCoins{sdk.NewDecCoinFromDec(s.baseDenom, commAmt)}
				// set outstanding rewards
				s.Require().NoError(s.network.App.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: valCommission}))
				// set commission
				s.Require().NoError(s.network.App.DistrKeeper.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: valCommission}))

				// set distribution module account balance which pays out the rewards
				coins := sdk.NewCoins(sdk.NewCoin(s.bondDenom, commAmt.RoundInt()))
				err = s.mintCoinsForDistrMod(ctx, coins)
				s.Require().NoError(err, "failed to fund distr module account")

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
				val := s.network.GetValidators()[0]
				ctx, err = s.prepareStakingRewards(
					ctx,
					stakingRewards{
						Delegator: s.keyring.GetAccAddr(0),
						Validator: val,
						RewardAmt: testRewardsAmt,
					},
				)
				s.Require().NoError(err, "failed to prepare staking rewards")

				input, err := s.precompile.Pack(
					distribution.WithdrawDelegatorRewardsMethod,
					s.keyring.GetAddr(0),
					val.OperatorAddress,
				)
				s.Require().NoError(err, "failed to pack input")

				return s.keyring.GetAddr(0), input
			},
			readOnly: false,
			expPass:  true,
		},
		{
			name: "pass - claim rewards transaction",
			malleate: func() (common.Address, []byte) {
				ctx, err = s.prepareStakingRewards(
					ctx,
					stakingRewards{
						Delegator: s.keyring.GetAccAddr(0),
						Validator: s.network.GetValidators()[0],
						RewardAmt: testRewardsAmt,
					},
				)
				s.Require().NoError(err, "failed to prepare staking rewards")

				input, err := s.precompile.Pack(
					distribution.ClaimRewardsMethod,
					s.keyring.GetAddr(0),
					uint32(2),
				)
				s.Require().NoError(err, "failed to pack input")

				return s.keyring.GetAddr(0), input
			},
			readOnly: false,
			expPass:  true,
		},
		{
			name: "pass - fund community pool transaction",
			malleate: func() (common.Address, []byte) {
				input, err := s.precompile.Pack(
					distribution.FundCommunityPoolMethod,
					s.keyring.GetAddr(0),
					big.NewInt(1e18),
				)
				s.Require().NoError(err, "failed to pack input")

				return s.keyring.GetAddr(0), input
			},
			readOnly: false,
			expPass:  true,
		},
		{
			name: "pass - fund community pool transaction",
			malleate: func() (common.Address, []byte) {
				input, err := s.precompile.Pack(
					distribution.FundCommunityPoolMethod,
					s.keyring.GetAddr(0),
					big.NewInt(1e18),
				)
				s.Require().NoError(err, "failed to pack input")

				return s.keyring.GetAddr(0), input
			},
			readOnly: false,
			expPass:  true,
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			// setup basic test suite
			s.SetupTest()
			ctx = s.network.GetContext()
			baseFee := s.network.App.EvmKeeper.GetBaseFee(ctx)

			// malleate testcase
			caller, input := tc.malleate()

			contract := vm.NewPrecompile(vm.AccountRef(caller), s.precompile, big.NewInt(0), uint64(1e6))
			contract.Input = input

			contractAddr := contract.Address()

			evmChainID := evmtypes.GetEthChainConfig().ChainID
			// Build and sign Ethereum transaction
			txArgs := evmtypes.EvmTxArgs{
				ChainID:   evmChainID,
				Nonce:     0,
				To:        &contractAddr,
				Amount:    nil,
				GasLimit:  100000,
				GasPrice:  app.MainnetMinGasPrices.BigInt(),
				GasFeeCap: baseFee,
				GasTipCap: big.NewInt(1),
				Accesses:  &gethtypes.AccessList{},
			}
			msgEthereumTx, err := s.factory.GenerateMsgEthereumTx(s.keyring.GetPrivKey(0), txArgs)
			s.Require().NoError(err, "failed to generate Ethereum message")

			signedMsg, err := s.factory.SignMsgEthereumTx(s.keyring.GetPrivKey(0), msgEthereumTx)
			s.Require().NoError(err, "failed to sign Ethereum message")

			// Instantiate config
			proposerAddress := ctx.BlockHeader().ProposerAddress
			cfg, err := s.network.App.EvmKeeper.EVMConfig(ctx, proposerAddress)
			s.Require().NoError(err, "failed to instantiate EVM config")

			ethChainID := s.network.GetEIP155ChainID()
			signer := gethtypes.LatestSignerForChainID(ethChainID)
			msg, err := signedMsg.AsMessage(signer, baseFee)
			s.Require().NoError(err, "failed to instantiate Ethereum message")

			// Instantiate EVM
			evm := s.network.App.EvmKeeper.NewEVM(
				ctx, msg, cfg, nil, s.network.GetStateDB(),
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
			}
		})
	}
}
