package staking_test

import (
	"fmt"
	"math/big"
	"time"

	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"

	"cosmossdk.io/math"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkauthz "github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/evmos/evmos/v20/precompiles/authorization"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/staking"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	"github.com/evmos/evmos/v20/x/evm/statedb"
)

func (s *PrecompileTestSuite) TestApprove() {
	var (
		ctx  sdk.Context
		stDB *statedb.StateDB
	)
	method := s.precompile.Methods[authorization.ApproveMethod]

	testCases := []struct {
		name        string
		malleate    func(contract *vm.Contract, granter, grantee testkeyring.Key) []interface{}
		postCheck   func(granter, grantee testkeyring.Key, data []byte, inputArgs []interface{})
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func(_ *vm.Contract, _, _ testkeyring.Key) []interface{} {
				return []interface{}{}
			},
			func(_, _ testkeyring.Key, _ []byte, _ []interface{}) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
		},
		{
			"fail - invalid message type",
			func(_ *vm.Contract, _, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					abi.MaxUint256,
					[]string{"invalid"},
				}
			},
			func(_, _ testkeyring.Key, _ []byte, _ []interface{}) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidMsgType, "staking", "invalid"),
		},
		// TODO: enable this test once we check if spender and origin are the same
		// {
		//	"fail - origin address is the same the spender address",
		//	func(_ *vm.Contract) []interface{} {
		//		return []interface{}{
		//			s.keyring.GetAddr(0),
		//			abi.MaxUint256,
		//			[]string{"invalid"},
		//		}
		//	},
		//	(data []byte inputArgs []interface{}) {},
		//	200000,
		//	true,
		//	"is the same as spender",
		// },
		{
			"success - MsgDelegate with unlimited coins",
			func(_ *vm.Contract, _, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					abi.MaxUint256,
					[]string{staking.DelegateMsg},
				}
			},
			func(granter, grantee testkeyring.Key, data []byte, _ []interface{}) {
				s.Require().Equal(data, cmn.TrueValue)
				authz, expirationTime := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)

				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)
				s.Require().Equal(authz.AuthorizationType, staking.DelegateAuthz)
				var coin *sdk.Coin
				s.Require().Equal(authz.MaxTokens, coin)
			},
			20000,
			false,
			"",
		},
		{
			"success - MsgUndelegate with unlimited coins",
			func(_ *vm.Contract, _, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					abi.MaxUint256,
					[]string{staking.UndelegateMsg},
				}
			},
			func(granter, grantee testkeyring.Key, data []byte, _ []interface{}) {
				s.Require().Equal(data, cmn.TrueValue)

				authz, expirationTime := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.UndelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)
				s.Require().Equal(authz.AuthorizationType, staking.UndelegateAuthz)
				var coin *sdk.Coin
				s.Require().Equal(authz.MaxTokens, coin)
			},
			20000,
			false,
			"",
		},
		{
			"success - MsgRedelegate with unlimited coins",
			func(_ *vm.Contract, _, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					abi.MaxUint256,
					[]string{staking.RedelegateMsg},
				}
			},
			func(granter, grantee testkeyring.Key, data []byte, _ []interface{}) {
				s.Require().Equal(data, cmn.TrueValue)

				authz, expirationTime := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.RedelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)
				s.Require().Equal(authz.AuthorizationType, staking.RedelegateAuthz)
				var coin *sdk.Coin
				s.Require().Equal(authz.MaxTokens, coin)
			},
			20000,
			false,
			"",
		},
		{
			"success - All staking methods with certain amount of coins",
			func(_ *vm.Contract, _, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					big.NewInt(1e18),
					[]string{
						staking.DelegateMsg,
						staking.UndelegateMsg,
						staking.RedelegateMsg,
					},
				}
			},
			func(granter, grantee testkeyring.Key, data []byte, _ []interface{}) {
				s.Require().Equal(data, cmn.TrueValue)

				allAuthz, err := s.network.App.AuthzKeeper.GetAuthorizations(ctx, grantee.AccAddr, granter.AccAddr)
				s.Require().NoError(err)
				s.Require().Len(allAuthz, 3)
			},
			20000,
			false,
			"",
		},
		{
			"success - remove MsgDelegate authorization",
			func(_ *vm.Contract, granter, grantee testkeyring.Key) []interface{} {
				res, err := s.precompile.Approve(ctx, granter.Addr, stDB, &method, []interface{}{
					grantee.Addr, big.NewInt(1), []string{staking.DelegateMsg},
				})
				s.Require().NoError(err)
				s.Require().Equal(res, cmn.TrueValue)

				authz, expirationTime := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)

				return []interface{}{
					grantee.Addr,
					big.NewInt(0),
					[]string{staking.DelegateMsg},
				}
			},
			func(granter, grantee testkeyring.Key, data []byte, _ []interface{}) {
				s.Require().Equal(data, cmn.TrueValue)

				authz, expirationTime := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().Nil(authz)
				s.Require().Nil(expirationTime)
			},
			200000,
			false,
			"",
		},
		{ //nolint:dupl
			"success - MsgDelegate with 1 Evmos as limit amount",
			func(_ *vm.Contract, _, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					big.NewInt(1e18),
					[]string{staking.DelegateMsg},
				}
			},
			func(granter, grantee testkeyring.Key, data []byte, _ []interface{}) {
				s.Require().Equal(data, cmn.TrueValue)

				authz, expirationTime := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)
				s.Require().Equal(authz.AuthorizationType, staking.DelegateAuthz)
				s.Require().Equal(authz.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)})
			},
			20000,
			false,
			"",
		},
		{
			"success - Authorization should only be created for validators that are not jailed",
			func(_ *vm.Contract, granter, grantee testkeyring.Key) []interface{} {
				var err error
				// Jail a validator
				valAddr, err = sdk.ValAddressFromBech32(s.network.GetValidators()[0].GetOperator())
				s.Require().NoError(err)
				err = s.network.App.StakingKeeper.Jail(ctx, sdk.ConsAddress(valAddr))
				s.Require().NoError(err, "failed to jail a validator")

				// When a delegator redelegates/undelegates from a validator, the validator
				// switches to Unbonding status.
				// Thus, validators with this status should be considered for the authorization

				// Unbond another validator
				valAddr1, err := sdk.ValAddressFromBech32(s.network.GetValidators()[1].GetOperator())
				s.Require().NoError(err)
				amount, err := s.network.App.StakingKeeper.Unbond(ctx, granter.AccAddr, valAddr1, math.LegacyOneDec())
				s.Require().NoError(err, "expected no error unbonding validator")
				s.Require().Equal(math.NewInt(1e18), amount, "expected different amount of tokens to be unbonded")

				return []interface{}{
					grantee.Addr,
					big.NewInt(1e18),
					[]string{staking.DelegateMsg},
				}
			},
			func(granter, grantee testkeyring.Key, data []byte, _ []interface{}) {
				s.Require().Equal(data, cmn.TrueValue)

				authz, expirationTime := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)
				s.Require().Equal(authz.AuthorizationType, staking.DelegateAuthz)
				s.Require().Equal(authz.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)})
				// Check that the bonded and unbonding validators are included on the authorization
				s.Require().Len(authz.GetAllowList().Address, 2, "should only be two validators in the allow list")
			},
			1e6,
			false,
			"",
		},
		{ //nolint:dupl
			"success - MsgUndelegate with 1 Evmos as limit amount",
			func(_ *vm.Contract, _, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					big.NewInt(1e18),
					[]string{staking.UndelegateMsg},
				}
			},
			func(granter, grantee testkeyring.Key, data []byte, _ []interface{}) {
				s.Require().Equal(data, cmn.TrueValue)

				authz, expirationTime := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.UndelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)
				s.Require().Equal(authz.AuthorizationType, staking.UndelegateAuthz)
				s.Require().Equal(authz.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)})
			},
			20000,
			false,
			"",
		},
		{
			"success - MsgRedelegate with 1 Evmos as limit amount",
			func(_ *vm.Contract, _, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					big.NewInt(1e18),
					[]string{staking.RedelegateMsg},
				}
			},
			func(granter, grantee testkeyring.Key, data []byte, _ []interface{}) {
				s.Require().Equal(data, cmn.TrueValue)

				authz, expirationTime := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.RedelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)
			},
			20000,
			false,
			"",
		},
		{
			"success - MsgRedelegate, MsgUndelegate and MsgDelegate with 1 Evmos as limit amount",
			func(_ *vm.Contract, _, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					big.NewInt(1e18),
					[]string{
						staking.RedelegateMsg,
						staking.UndelegateMsg,
						staking.DelegateMsg,
					},
				}
			},
			func(granter, grantee testkeyring.Key, data []byte, _ []interface{}) {
				s.Require().Equal(data, cmn.TrueValue)

				authz, expirationTime := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)
				s.Require().Equal(authz.AuthorizationType, staking.DelegateAuthz)
				s.Require().Equal(authz.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)})

				authz, expirationTime = CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.UndelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)

				s.Require().Equal(authz.AuthorizationType, staking.UndelegateAuthz)
				s.Require().Equal(authz.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)})

				authz, expirationTime = CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.RedelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().NotNil(expirationTime)

				s.Require().Equal(authz.AuthorizationType, staking.RedelegateAuthz)
				s.Require().Equal(authz.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)})

				// TODO: Bug here it returns 3 REDELEGATE authorizations
				allAuthz, err := s.network.App.AuthzKeeper.GetAuthorizations(s.network.GetContext(), grantee.AccAddr, granter.AccAddr)
				s.Require().NoError(err)
				s.Require().Len(allAuthz, 3)
			},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases { //nolint:dupl
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, granter.Addr, s.precompile, tc.gas)

			args := tc.malleate(contract, granter, grantee)
			bz, err := s.precompile.Approve(ctx, granter.Addr, stDB, &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(granter, grantee, bz, args)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestDecreaseAllowance() {
	var (
		ctx  sdk.Context
		stDB *statedb.StateDB
	)
	method := s.precompile.Methods[authorization.DecreaseAllowanceMethod]

	testCases := []struct {
		name        string
		malleate    func(_ *vm.Contract, granter, grantee testkeyring.Key) []interface{}
		postCheck   func(granter, grantee testkeyring.Key, data []byte, inputArgs []interface{})
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func(_ *vm.Contract, _, _ testkeyring.Key) []interface{} {
				return []interface{}{}
			},
			func(_, _ testkeyring.Key, _ []byte, _ []interface{}) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
		},
		//// TODO: enable this once we check origin is not the spender
		// {
		//	"fail - origin address is the spender address",
		//	func(_ *vm.Contract) []interface{} {
		//		return []interface{}{
		//			s.keyring.GetAddr(0),
		//			abi.MaxUint256,
		//			[]string{staking.DelegateMsg},
		//		}
		//	},
		//	(data []byte inputArgs []interface{}) {},
		//	200000,
		//	true,
		//	"is the same as spender",
		// },
		{
			"fail - delegate authorization does not exists",
			func(_ *vm.Contract, _, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					big.NewInt(15000),
					[]string{staking.DelegateMsg},
				}
			},
			func(_, _ testkeyring.Key, _ []byte, _ []interface{}) {
			},
			200000,
			true,
			"authorization to /cosmos.staking.v1beta1.MsgDelegate",
		},
		{
			"fail - delegate authorization is a generic Authorization",
			func(_ *vm.Contract, granter, grantee testkeyring.Key) []interface{} {
				authz := sdkauthz.NewGenericAuthorization(staking.DelegateMsg)
				exp := time.Now().Add(time.Hour)
				err := s.network.App.AuthzKeeper.SaveGrant(ctx, grantee.AccAddr, granter.AccAddr, authz, &exp)
				s.Require().NoError(err)
				return []interface{}{
					grantee.Addr,
					big.NewInt(15000),
					[]string{staking.DelegateMsg},
				}
			},
			func(_, _ testkeyring.Key, _ []byte, _ []interface{}) {
			},
			200000,
			true,
			sdkauthz.ErrUnknownAuthorizationType.Error(),
		},
		{
			"fail - decrease allowance amount is greater than the authorization limit",
			func(_ *vm.Contract, granter, grantee testkeyring.Key) []interface{} {
				approveArgs := []interface{}{
					grantee.Addr,
					big.NewInt(1e18),
					[]string{staking.DelegateMsg},
				}
				resp, err := s.precompile.Approve(ctx, granter.Addr, stDB, &method, approveArgs)
				s.Require().NoError(err)
				s.Require().Equal(resp, cmn.TrueValue)

				authz, _ := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().Equal(authz.AuthorizationType, staking.DelegateAuthz)
				s.Require().Equal(authz.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)})

				return []interface{}{
					grantee.Addr,
					big.NewInt(2e18),
					[]string{staking.DelegateMsg},
				}
			},
			func(_, _ testkeyring.Key, _ []byte, _ []interface{}) {},
			200000,
			true,
			"amount by which the allowance should be decreased is greater than the authorization limit",
		},
		{
			"success - decrease delegate authorization allowance by 1 Evmos",
			func(_ *vm.Contract, granter, grantee testkeyring.Key) []interface{} {
				s.ApproveAndCheckAuthz(method, granter, grantee, staking.DelegateMsg, big.NewInt(2e18))
				return []interface{}{
					grantee.Addr,
					big.NewInt(1e18),
					[]string{staking.DelegateMsg},
				}
			},
			func(granter, grantee testkeyring.Key, _ []byte, _ []interface{}) {
				authz, _ := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().Equal(authz.AuthorizationType, staking.DelegateAuthz)
				s.Require().Equal(authz.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)})
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases { //nolint:dupl
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			var contract *vm.Contract
			contract, ctx = testutil.NewPrecompileContract(s.T(), ctx, granter.Addr, s.precompile, tc.gas)

			args := tc.malleate(contract, granter, grantee)
			bz, err := s.precompile.DecreaseAllowance(ctx, granter.Addr, stDB, &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(granter, grantee, bz, args)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestIncreaseAllowance() {
	var (
		ctx  sdk.Context
		stDB *statedb.StateDB
	)
	method := s.precompile.Methods[authorization.IncreaseAllowanceMethod]

	testCases := []struct {
		name        string
		malleate    func(granter, grantee testkeyring.Key) []interface{}
		postCheck   func(granter, grantee testkeyring.Key, data []byte, inputArgs []interface{})
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func(_, _ testkeyring.Key) []interface{} {
				return []interface{}{}
			},
			func(_, _ testkeyring.Key, _ []byte, _ []interface{}) {},
			200000,
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 3, 0),
		},
		// TODO: enable this once we check origin is not the same as spender
		// {
		//	"fail - origin address is the spender address",
		//	func(_ *vm.Contract) []interface{} {
		//		return []interface{}{
		//			s.keyring.GetAddr(0),
		//			abi.MaxUint256,
		//			[]string{staking.DelegateMsg},
		//		}
		//	},
		//	(data []byte inputArgs []interface{}) {},
		//	200000,
		//	true,
		//	"is the same as spender",
		// },
		{
			"fail - delegate authorization does not exists",
			func(_, grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					big.NewInt(15000),
					[]string{staking.DelegateMsg},
				}
			},
			func(_, _ testkeyring.Key, _ []byte, _ []interface{}) {
			},
			200000,
			true,
			"authorization to /cosmos.staking.v1beta1.MsgDelegate",
		},
		{
			"success - no-op, allowance amount is already set to the maximum value",
			func(granter, grantee testkeyring.Key) []interface{} {
				approveArgs := []interface{}{
					grantee.Addr,
					abi.MaxUint256,
					[]string{staking.DelegateMsg},
				}
				resp, err := s.precompile.Approve(ctx, granter.Addr, stDB, &method, approveArgs)
				s.Require().NoError(err)
				s.Require().Equal(resp, cmn.TrueValue)

				authz, _ := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().Equal(authz.AuthorizationType, staking.DelegateAuthz)
				var coin *sdk.Coin
				s.Require().Equal(authz.MaxTokens, coin)

				return []interface{}{
					grantee.Addr,
					big.NewInt(2e18),
					[]string{staking.DelegateMsg},
				}
			},
			func(_, _ testkeyring.Key, _ []byte, _ []interface{}) {},
			200000,
			false,
			"",
		},
		{
			"success - increase delegate authorization allowance by 1 Evmos",
			func(granter, grantee testkeyring.Key) []interface{} {
				s.ApproveAndCheckAuthz(method, granter, grantee, staking.DelegateMsg, big.NewInt(1e18))
				return []interface{}{
					grantee.Addr,
					big.NewInt(1e18),
					[]string{staking.DelegateMsg},
				}
			},
			func(granter, grantee testkeyring.Key, _ []byte, _ []interface{}) {
				authz, _ := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
				s.Require().Equal(authz.AuthorizationType, staking.DelegateAuthz)
				s.Require().Equal(authz.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(2e18)})
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()
			stDB = s.network.GetStateDB()

			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			args := tc.malleate(granter, grantee)
			bz, err := s.precompile.IncreaseAllowance(ctx, granter.Addr, stDB, &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(granter, grantee, bz, args)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestRevoke() {
	var ctx sdk.Context

	method := s.precompile.Methods[authorization.RevokeMethod]
	createdAuthz := staking.DelegateAuthz
	approvedCoin := &sdk.Coin{Denom: s.bondDenom, Amount: math.NewInt(1e18)}

	testCases := []struct {
		name        string
		malleate    func(grantee testkeyring.Key) []interface{}
		postCheck   func(granter, grantee testkeyring.Key, data []byte, inputArgs []interface{})
		expError    bool
		errContains string
	}{
		{
			name: "fail - empty input args",
			malleate: func(_ testkeyring.Key) []interface{} {
				return []interface{}{}
			},
			expError:    true,
			errContains: fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		},
		{
			name: "fail - authorization does not exist",
			malleate: func(grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					[]string{staking.UndelegateMsg},
				}
			},
			postCheck: func(granter, grantee testkeyring.Key, _ []byte, _ []interface{}) {
				// expect authorization to still be there
				authz, _ := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, createdAuthz, grantee.Addr, granter.Addr)
				s.Require().NotNil(authz)
			},
			expError:    true,
			errContains: "authorization not found",
		},
		{
			name: "pass - authorization revoked",
			malleate: func(grantee testkeyring.Key) []interface{} {
				return []interface{}{
					grantee.Addr,
					[]string{staking.DelegateMsg},
				}
			},
			postCheck: func(granter, grantee testkeyring.Key, _ []byte, _ []interface{}) {
				// expect authorization to be removed
				authz, _ := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, createdAuthz, grantee.Addr, granter.Addr)
				s.Require().Nil(authz, "expected authorization to be removed")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset
			ctx = s.network.GetContext()

			granter := s.keyring.GetKey(0)
			grantee := s.keyring.GetKey(1)

			// Create a delegate authorization
			validators, err := s.network.App.StakingKeeper.GetLastValidators(ctx)
			s.Require().NoError(err)
			valAddrs := make([]sdk.ValAddress, len(validators))
			for i, val := range validators {
				valAddrs[i] = sdk.ValAddress(val.GetOperator())
			}
			delegationAuthz, err := stakingtypes.NewStakeAuthorization(
				valAddrs,
				nil,
				createdAuthz,
				approvedCoin,
			)
			s.Require().NoError(err)

			expiration := ctx.BlockTime().Add(time.Hour * 24 * 365).UTC()
			err = s.network.App.AuthzKeeper.SaveGrant(ctx, grantee.AccAddr, granter.AccAddr, delegationAuthz, &expiration)
			s.Require().NoError(err, "failed to save authorization")
			authz, _ := CheckAuthorizationWithContext(ctx, s.network.App.AuthzKeeper, createdAuthz, grantee.Addr, granter.Addr)
			s.Require().NotNil(authz, "expected authorization to be set")

			args := tc.malleate(grantee)
			bz, err := s.precompile.Revoke(ctx, granter.Addr, s.network.GetStateDB(), &method, args)

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
				s.Require().Empty(bz)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(granter, grantee, bz, args)
			}
		})
	}
}
