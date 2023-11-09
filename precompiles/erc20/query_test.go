// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20_test

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/evmos/evmos/v15/app"
	"github.com/evmos/evmos/v15/precompiles/erc20"
	inflationtypes "github.com/evmos/evmos/v15/x/inflation/types"
)

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
)

// TestNameSymbol tests the Name and Symbol methods of the ERC20 precompile.
//
// NOTE: we test both methods in the same test because they are need the same testcases and
// the same setup.
func (s *PrecompileTestSuite) TestNameSymbol() {
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
			name:                "fail - empty denom",
			denom:               "",
			errContains:         "denom cannot be empty",
			errDecimalsContains: "denom is not an IBC voucher", // TODO: do we want to check for empty denom here too?
		},
		{
			name:                "fail - invalid denom trace",
			denom:               tooShortTrace.IBCDenom()[:len(tooShortTrace.IBCDenom())-1],
			errContains:         "odd length hex string",
			errDecimalsContains: "odd length hex string",
		},
		{
			name:                "fail - denom not found",
			denom:               types.DenomTrace{Path: "channel-0", BaseDenom: "notfound"}.IBCDenom(),
			errContains:         "denom trace not found",
			errDecimalsContains: "denom trace not found",
		},
		{
			name:  "fail - invalid denom (too short < 3 chars)",
			denom: tooShortTrace.IBCDenom(),
			malleate: func(ctx sdk.Context, app *app.Evmos) {
				app.TransferKeeper.SetDenomTrace(ctx, tooShortTrace)
			},
			errContains:     "invalid base denomination; should be at least length 3; got: \"ab\"",
			expDecimalsPass: true, // TODO: do we want to check in decimals query for the above error?
			expDecimals:     18,   // expect 18 decimals here because of "a" prefix
		},
		{
			name:                "fail - denom without metadata and not an IBC voucher",
			denom:               "noIBCvoucher",
			errContains:         "denom is not an IBC voucher",
			errDecimalsContains: "denom is not an IBC voucher",
		},
		{
			name:  "fail - valid ibc denom without metadata and neither atto nor micro prefix",
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
			name:  "pass - valid denom with metadata",
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
			name:  "pass - valid ibc denom without metadata",
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
			name:  "pass - valid ibc denom with metadata and 18 decimals",
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
			name:  "fail - valid ibc denom with metadata but decimals overflow",
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
	}

	for _, tc := range testcases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			if tc.malleate != nil {
				tc.malleate(s.network.GetContext(), s.network.App)
			}

			precompile, _ := s.setupERC20Precompile(tc.denom)

			s.Run("name", func() {
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

			s.Run("symbol", func() {
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

			s.Run("decimals", func() {
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
