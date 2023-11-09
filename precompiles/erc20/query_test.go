// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20_test

import (
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/app"
	auth "github.com/evmos/evmos/v15/precompiles/authorization"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	"github.com/evmos/evmos/v15/testutil"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/v1/types"
)

// Define useful variables for tests here.
var (
	// tooShortTrace is a denomination trace with a name that will raise the "denom too short" error
	tooShortTrace = types.DenomTrace{Path: "channel-0", BaseDenom: "ab"}
	// validTraceDenom is a denomination trace with a valid IBC voucher name
	validTraceDenom = types.DenomTrace{Path: "channel-0", BaseDenom: "uosmo"}
	// validAttoTraceDenom is a denomination trace with a valid IBC voucher name and 18 decimals
	validAttoTraceDenom = types.DenomTrace{Path: "channel-0", BaseDenom: "aevmos"}
	// validTraceDenomNoMicroAtto is a denomination trace with a valid IBC voucher name but no micro or atto prefix
	validTraceDenomNoMicroAtto = types.DenomTrace{Path: "channel-0", BaseDenom: "mevmos"}

	// --------------------
	// Variables for coin with valid metadata
	//

	// validMetadataDenom is the base denomination of the coin with valid metadata
	validMetadataDenom = "uatom"
	// validMetadataDisplay is the denomination displayed of the coin with valid metadata
	validMetadataDisplay = "atom"
	// validMetadataName is the name of the coin with valid metadata
	validMetadataName = "Atom"
	// validMetadataSymbol is the symbol of the coin with valid metadata
	validMetadataSymbol = "ATOM"

	// validMetadata is the metadata of the coin with valid metadata
	validMetadata = banktypes.Metadata{
		Description: "description",
		Base:        validMetadataDenom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    validMetadataDenom,
				Exponent: 0,
			},
			{
				Denom:    validMetadataDisplay,
				Exponent: uint32(6),
			},
		},
		Name:    validMetadataName,
		Symbol:  validMetadataSymbol,
		Display: validMetadataDisplay,
	}

	// overflowMetadata contains a metadata with an exponent that overflows uint8
	overflowMetadata = banktypes.Metadata{
		Description: "description",
		Base:        validMetadataDenom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    validMetadataDenom,
				Exponent: 0,
			},
			{
				Denom:    validMetadataDisplay,
				Exponent: uint32(math.MaxUint8 + 1),
			},
		},
		Name:    validMetadataName,
		Symbol:  validMetadataSymbol,
		Display: validMetadataDisplay,
	}

	// noDisplayMetadata contains a metadata where the denom units do not contain with no display denom
	noDisplayMetadata = banktypes.Metadata{
		Description: "description",
		Base:        validMetadataDenom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    validMetadataDenom,
				Exponent: 0,
			},
		},
		Name:    validMetadataName,
		Symbol:  validMetadataSymbol,
		Display: "",
	}
)

// TestNameSymbolDecimals tests the Name, Symbol and Decimals methods of the ERC20 precompile.
//
// NOTE: we test both methods in the same test because they are need the same testcases and
// the same setup.
func (s *PrecompileTestSuite) TestNameSymbolDecimals() {
	nameMethod := s.precompile.Methods[erc20.NameMethod]
	symbolMethod := s.precompile.Methods[erc20.SymbolMethod]
	DecimalsMethod := s.precompile.Methods[erc20.DecimalsMethod]

	testcases := []struct {
		name                string
		denom               string
		malleate            func(sdk.Context, *app.Evmos)
		expPass             bool
		errContains         string
		expDecimalsPass     bool
		errDecimalsContains string
		expName             string
		expSymbol           string
		expDecimals         uint8
	}{
		{
			name:                "empty denom",
			denom:               "",
			errContains:         "denom is not an IBC voucher",
			errDecimalsContains: "denom is not an IBC voucher",
		},
		{
			name:                "invalid denom trace",
			denom:               tooShortTrace.IBCDenom()[:len(tooShortTrace.IBCDenom())-1],
			errContains:         "odd length hex string",
			errDecimalsContains: "odd length hex string",
		},
		{
			name:                "denom not found",
			denom:               types.DenomTrace{Path: "channel-0", BaseDenom: "notfound"}.IBCDenom(),
			errContains:         "denom trace not found",
			errDecimalsContains: "denom trace not found",
		},
		{
			name:  "invalid denom (too short < 3 chars)",
			denom: tooShortTrace.IBCDenom(),
			malleate: func(ctx sdk.Context, app *app.Evmos) {
				app.TransferKeeper.SetDenomTrace(ctx, tooShortTrace)
			},
			errContains:     "invalid base denomination; should be at least length 3; got: \"ab\"",
			expDecimalsPass: true, // TODO: do we want to check in decimals query for the above error?
			expDecimals:     18,   // expect 18 decimals here because of "a" prefix
		},
		{
			name:                "denom without metadata and not an IBC voucher",
			denom:               "noIBCvoucher",
			errContains:         "denom is not an IBC voucher",
			errDecimalsContains: "denom is not an IBC voucher",
		},
		{
			name:  "valid ibc denom without metadata and neither atto nor micro prefix",
			denom: validTraceDenomNoMicroAtto.IBCDenom(),
			malleate: func(ctx sdk.Context, app *app.Evmos) {
				app.TransferKeeper.SetDenomTrace(ctx, validTraceDenomNoMicroAtto)
			},
			expPass:             true,
			expName:             "Evmos",
			expSymbol:           "EVMOS",
			errDecimalsContains: "invalid base denomination; should be either micro ('u[...]') or atto ('a[...]'); got: \"mevmos\"",
		},
		{
			name:  "valid denom with metadata",
			denom: validMetadataDenom,
			malleate: func(ctx sdk.Context, app *app.Evmos) {
				// NOTE: we mint some coins to the inflation module address to be able to set denom metadata
				err := app.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(validMetadata.Base, 1)})
				s.Require().NoError(err)

				// NOTE: we set the denom metadata for the coin
				app.BankKeeper.SetDenomMetaData(ctx, validMetadata)
			},
			expPass:         true,
			expDecimalsPass: true,
			expName:         "Atom",
			expSymbol:       "ATOM",
			expDecimals:     6,
		},
		{
			name:  "valid ibc denom without metadata",
			denom: validTraceDenom.IBCDenom(),
			malleate: func(ctx sdk.Context, app *app.Evmos) {
				app.TransferKeeper.SetDenomTrace(ctx, validTraceDenom)
			},
			expPass:         true,
			expDecimalsPass: true,
			expName:         "Osmo",
			expSymbol:       "OSMO",
			expDecimals:     6,
		},
		{
			name:  "valid ibc denom without metadata and 18 decimals",
			denom: validAttoTraceDenom.IBCDenom(),
			malleate: func(ctx sdk.Context, app *app.Evmos) {
				app.TransferKeeper.SetDenomTrace(ctx, validAttoTraceDenom)
			},
			expPass:         true,
			expDecimalsPass: true,
			expName:         "Evmos",
			expSymbol:       "EVMOS",
			expDecimals:     18,
		},
		{
			name:  "valid ibc denom with metadata but decimals overflow",
			denom: validMetadataDenom,
			malleate: func(ctx sdk.Context, app *app.Evmos) {
				// NOTE: we mint some coins to the inflation module address to be able to set denom metadata
				err := app.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(validMetadata.Base, 1)})
				s.Require().NoError(err)

				// NOTE: we set the denom metadata for the coin
				app.BankKeeper.SetDenomMetaData(s.network.GetContext(), overflowMetadata)
			},
			expPass:             true,
			expDecimalsPass:     false,
			expName:             "Atom",
			expSymbol:           "ATOM",
			errDecimalsContains: "uint8 overflow: invalid decimals",
		},
		{
			name:  "valid ibc denom with metadata but no display denom",
			denom: validMetadataDenom,
			malleate: func(ctx sdk.Context, app *app.Evmos) {
				// NOTE: we mint some coins to the inflation module address to be able to set denom metadata
				err := app.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(validMetadata.Base, 1)})
				s.Require().NoError(err)

				// NOTE: we set the denom metadata for the coin
				app.BankKeeper.SetDenomMetaData(ctx, noDisplayMetadata)
			},
			expPass:             true,
			expDecimalsPass:     false,
			expName:             "Atom",
			expSymbol:           "ATOM",
			errDecimalsContains: "display denomination not found for denom: \"uatom\"",
		},
	}

	for _, tc := range testcases {
		tc := tc

		testNameSymbolPostfix := " - pass"
		if !tc.expPass {
			testNameSymbolPostfix = " - fail"
		}
		testDecimalsPostfix := " - pass"
		if !tc.expDecimalsPass {
			testDecimalsPostfix = " - fail"
		}

		s.Run(tc.name, func() {
			s.SetupTest()

			if tc.malleate != nil {
				tc.malleate(s.network.GetContext(), s.network.App)
			}

			precompile := s.setupERC20Precompile(tc.denom)

			s.Run("name"+testNameSymbolPostfix, func() {
				bz, err := precompile.Name(
					s.network.GetContext(),
					nil,
					nil,
					&nameMethod,
					[]interface{}{},
				)

				if tc.expPass {
					s.Require().NoError(err, "expected no error getting name")
					s.Require().NotEmpty(bz, "expected name bytes not to be empty")

					// Unpack the name into a string
					nameOut, err := nameMethod.Outputs.Unpack(bz)
					s.Require().NoError(err, "expected no error unpacking name")
					s.Require().Equal(tc.expName, nameOut[0], "expected different name")
				} else {
					s.Require().Error(err, "expected error getting name")
					s.Require().Contains(err.Error(), tc.errContains, "expected different error getting name")
				}
			})

			s.Run("symbol"+testNameSymbolPostfix, func() {
				bz, err := precompile.Symbol(
					s.network.GetContext(),
					nil,
					nil,
					&symbolMethod,
					[]interface{}{},
				)

				if tc.expPass {
					s.Require().NoError(err, "expected no error getting symbol")
					s.Require().NotEmpty(bz, "expected symbol bytes not to be empty")

					// Unpack the name into a string
					symbolOut, err := symbolMethod.Outputs.Unpack(bz)
					s.Require().NoError(err, "expected no error unpacking symbol")
					s.Require().Equal(tc.expSymbol, symbolOut[0], "expected different symbol")
				} else {
					s.Require().Error(err, "expected error getting symbol")
					s.Require().Contains(err.Error(), tc.errContains, "expected different error getting symbol")
				}
			})

			s.Run("decimals"+testDecimalsPostfix, func() {
				bz, err := precompile.Decimals(
					s.network.GetContext(),
					nil,
					nil,
					&DecimalsMethod,
					[]interface{}{},
				)

				if tc.expDecimalsPass {
					s.Require().NoError(err, "expected no error getting decimals")
					s.Require().NotEmpty(bz, "expected decimals bytes not to be empty")

					// Unpack the name into a string
					decimalsOut, err := DecimalsMethod.Outputs.Unpack(bz)
					s.Require().NoError(err, "expected no error unpacking decimals")
					s.Require().Equal(tc.expDecimals, decimalsOut[0], "expected different decimals")
				} else {
					s.Require().Error(err, "expected error getting decimals")
					s.Require().Contains(err.Error(), tc.errDecimalsContains, "expected different error getting decimals")
				}
			})
		})
	}
}

func (s *PrecompileTestSuite) TestTotalSupply() {
	method := s.precompile.Methods[erc20.TotalSupplyMethod]

	testcases := []struct {
		name        string
		malleate    func(sdk.Context, *app.Evmos, *big.Int)
		expPass     bool
		errContains string
		expTotal    *big.Int
	}{
		{
			name:     "pass - no coins",
			expPass:  true,
			expTotal: common.Big0,
		},
		{
			name: "pass - some coins",
			malleate: func(ctx sdk.Context, app *app.Evmos, amount *big.Int) {
				// NOTE: we mint some coins to the inflation module address to be able to set denom metadata
				err := app.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, sdk.Coins{sdk.NewCoin(validMetadata.Base, sdk.NewIntFromBigInt(amount))})
				s.Require().NoError(err)
			},
			expPass:  true,
			expTotal: big.NewInt(100),
		},
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			if tc.malleate != nil {
				tc.malleate(s.network.GetContext(), s.network.App, tc.expTotal)
			}

			precompile := s.setupERC20Precompile(validMetadataDenom)

			bz, err := precompile.TotalSupply(
				s.network.GetContext(),
				nil,
				nil,
				&method,
				[]interface{}{},
			)

			s.requireOut(bz, err, method, tc.expPass, tc.errContains, tc.expTotal)
		})
	}
}

func (s *PrecompileTestSuite) TestBalanceOf() {
	method := s.precompile.Methods[erc20.BalanceOfMethod]

	testcases := []struct {
		name        string
		malleate    func(sdk.Context, *app.Evmos, *big.Int) []interface{}
		expPass     bool
		errContains string
		expBalance  *big.Int
	}{
		{
			name: "fail - invalid number of arguments",
			malleate: func(_ sdk.Context, _ *app.Evmos, _ *big.Int) []interface{} {
				return []interface{}{}
			},
			errContains: "invalid number of arguments; expected 1; got: 0",
		},
		{
			name: "fail - invalid address",
			malleate: func(_ sdk.Context, _ *app.Evmos, _ *big.Int) []interface{} {
				return []interface{}{"invalid address"}
			},
			errContains: "invalid account address: invalid address",
		},
		{
			name: "pass - no coins in token denomination of precompile token pair",
			malleate: func(_ sdk.Context, _ *app.Evmos, _ *big.Int) []interface{} {
				// NOTE: we fund the account with some coins in a different denomination from what was used in the precompile.
				err := testutil.FundAccount(
					s.network.GetContext(), s.network.App.BankKeeper, s.keyring.GetAccAddr(0), sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, 100)),
				)
				s.Require().NoError(err, "expected no error funding account")

				return []interface{}{s.keyring.GetAddr(0)}
			},
			expPass:    true,
			expBalance: common.Big0,
		},
		{
			name: "pass - some coins",
			malleate: func(ctx sdk.Context, app *app.Evmos, amount *big.Int) []interface{} {
				// NOTE: we fund the account with some coins of the token denomination that was used for the precompile
				err := testutil.FundAccount(
					ctx, app.BankKeeper, s.keyring.GetAccAddr(0), sdk.NewCoins(sdk.NewCoin(s.tokenDenom, sdk.NewIntFromBigInt(amount))),
				)
				s.Require().NoError(err, "expected no error funding account")

				return []interface{}{s.keyring.GetAddr(0)}
			},
			expPass:    true,
			expBalance: big.NewInt(100),
		},
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			var balanceOfArgs []interface{}
			if tc.malleate != nil {
				balanceOfArgs = tc.malleate(s.network.GetContext(), s.network.App, tc.expBalance)
			}

			precompile := s.setupERC20Precompile(s.tokenDenom)

			bz, err := precompile.BalanceOf(
				s.network.GetContext(),
				nil,
				nil,
				&method,
				balanceOfArgs,
			)

			s.requireOut(bz, err, method, tc.expPass, tc.errContains, tc.expBalance)
		})
	}
}

func (s *PrecompileTestSuite) TestAllowance() {
	method := s.precompile.Methods[auth.AllowanceMethod]

	testcases := []struct {
		name        string
		malleate    func(sdk.Context, *app.Evmos, *big.Int) []interface{}
		expPass     bool
		errContains string
		expAllow    *big.Int
	}{
		{
			name: "fail - invalid number of arguments",
			malleate: func(_ sdk.Context, _ *app.Evmos, _ *big.Int) []interface{} {
				return []interface{}{1}
			},
			errContains: "invalid number of arguments; expected 2; got: 1",
		},
		{
			name: "fail - invalid owner address",
			malleate: func(_ sdk.Context, _ *app.Evmos, _ *big.Int) []interface{} {
				return []interface{}{"invalid address", s.keyring.GetAddr(1)}
			},
			errContains: "invalid owner address: invalid address",
		},
		{
			name: "fail - invalid spender address",
			malleate: func(_ sdk.Context, _ *app.Evmos, _ *big.Int) []interface{} {
				return []interface{}{s.keyring.GetAddr(0), "invalid address"}
			},
			errContains: "invalid spender address: invalid address",
		},
		{
			name: "fail - no allowance exists",
			malleate: func(_ sdk.Context, _ *app.Evmos, _ *big.Int) []interface{} {
				return []interface{}{s.keyring.GetAddr(0), s.keyring.GetAddr(1)}
			},
			errContains: "does not exist or is expired",
		},
		{
			name: "pass - allowance exists but not for precompile token pair denom",
			malleate: func(_ sdk.Context, _ *app.Evmos, _ *big.Int) []interface{} {
				granterIdx := 0
				granteeIdx := 1

				s.setupSendAuthz(
					s.keyring.GetAccAddr(granteeIdx),
					s.keyring.GetPrivKey(granterIdx),
					sdk.NewCoins(sdk.NewInt64Coin(s.bondDenom, 100)),
				)

				return []interface{}{s.keyring.GetAddr(granterIdx), s.keyring.GetAddr(granteeIdx)}
			},
			expPass:  true,
			expAllow: common.Big0,
		},
		{
			name: "pass - allowance exists for precompile token pair denom",
			malleate: func(ctx sdk.Context, app *app.Evmos, amount *big.Int) []interface{} {
				granterIdx := 0
				granteeIdx := 1

				s.setupSendAuthz(
					s.keyring.GetAccAddr(granteeIdx),
					s.keyring.GetPrivKey(granterIdx),
					sdk.NewCoins(sdk.NewCoin(s.tokenDenom, sdk.NewIntFromBigInt(amount))),
				)

				return []interface{}{s.keyring.GetAddr(granterIdx), s.keyring.GetAddr(granteeIdx)}
			},
			expPass:  true,
			expAllow: big.NewInt(100),
		},
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			var allowanceArgs []interface{}
			if tc.malleate != nil {
				allowanceArgs = tc.malleate(s.network.GetContext(), s.network.App, tc.expAllow)
			}

			precompile := s.setupERC20Precompile(s.tokenDenom)

			bz, err := precompile.Allowance(
				s.network.GetContext(),
				nil,
				nil,
				&method,
				allowanceArgs,
			)

			s.requireOut(bz, err, method, tc.expPass, tc.errContains, tc.expAllow)
		})
	}
}

// requireOut is a helper utility to reduce the amount of boilerplate code in the tests.
//
// It requires the output bytes and error to match the expected values. Additionally, the method outputs
// are unpacked and the first value is compared to the expected value.
//
// NOTE: It's sufficient to only check the first value because all methods in the ERC20 precompile only
// return a single value.
func (s *PrecompileTestSuite) requireOut(
	bz []byte,
	err error,
	method abi.Method,
	expPass bool,
	errContains string,
	expValue interface{},
) {
	if expPass {
		s.Require().NoError(err, "expected no error")
		s.Require().NotEmpty(bz, "expected bytes not to be empty")

		// Unpack the name into a string
		out, err := method.Outputs.Unpack(bz)
		s.Require().NoError(err, "expected no error unpacking")

		// Check if expValue is a big.Int. Because of a difference in uninitialized/empty values for big.Ints,
		// this comparison is often not working as expected, so we convert to Int64 here and compare those values.
		bigExp, ok := expValue.(*big.Int)
		if ok {
			bigOut, ok := out[0].(*big.Int)
			s.Require().True(ok, "expected output to be a big.Int")
			s.Require().Equal(bigExp.Int64(), bigOut.Int64(), "expected different value")
		} else {
			s.Require().Equal(expValue, out[0], "expected different value")
		}
	} else {
		s.Require().Error(err, "expected error")
		s.Require().Contains(err.Error(), errContains, "expected different error")
	}
}
