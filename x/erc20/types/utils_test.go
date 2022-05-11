package types

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

func TestSanitizeERC20Name(t *testing.T) {
	testCases := []struct {
		name         string
		erc20Name    string
		expErc20Name string
	}{
		{"name contains 'Special Characters'", "*Special _ []{}||*¼^%  &Token", "SpecialToken"},
		{"name contains 'Spaces'", "   Spaces   Token", "SpacesToken"},
		{"name contains 'Leading Numbers'", "12313213  Number     Coin", "NumberCoin"},
		{"name contains 'Numbers in the middle'", "  Other    Erc20 Coin ", "OtherErc20Coin"},
		{"name contains '/'", "USD/Coin", "USD/Coin"},
		{"name contains '/'", "/SlashCoin", "SlashCoin"},
		{"name contains '/'", "O/letter", "O/letter"},
		{"name contains '/'", "Ot/2letters", "Ot/2letters"},
		{"name contains '/'", "123/leadingslash", "leadingslash"},
		{"name contains '-'", "Dash-Coin", "Dash-Coin"},
		{"really long word", strings.Repeat("a", 150), strings.Repeat("a", 127)},
		{"single word name: Token", "Token", "Token"},
		{"single word name: Coin", "Coin", "Coin"},
	}

	for _, tc := range testCases {
		name := SanitizeERC20Name(tc.erc20Name)
		require.Equal(t, tc.expErc20Name, name, tc.name)
		err := sdk.ValidateDenom(name)
		require.NoError(t, err)
	}
}

func TestEqualMetadata(t *testing.T) {
	testCases := []struct {
		name      string
		metadataA banktypes.Metadata
		metadataB banktypes.Metadata
		expError  bool
	}{
		{
			"equal metadata",
			banktypes.Metadata{
				Base:        "aevmos",
				Display:     "evmos",
				Name:        "Evmos",
				Symbol:      "EVMOS",
				Description: "EVM, staking and governance denom of Evmos",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    "aevmos",
						Exponent: 0,
						Aliases:  []string{"atto evmos"},
					},
					{
						Denom:    "evmos",
						Exponent: 18,
					},
				},
			},
			banktypes.Metadata{
				Base:        "aevmos",
				Display:     "evmos",
				Name:        "Evmos",
				Symbol:      "EVMOS",
				Description: "EVM, staking and governance denom of Evmos",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    "aevmos",
						Exponent: 0,
						Aliases:  []string{"atto evmos"},
					},
					{
						Denom:    "evmos",
						Exponent: 18,
					},
				},
			},
			false,
		},
		{
			"different base field",
			banktypes.Metadata{
				Base: "aevmos",
			},
			banktypes.Metadata{
				Base: "taevmos",
			},
			true,
		},
		{
			"different denom units length",
			banktypes.Metadata{
				Base:        "aevmos",
				Display:     "evmos",
				Name:        "Evmos",
				Symbol:      "EVMOS",
				Description: "EVM, staking and governance denom of Evmos",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    "aevmos",
						Exponent: 0,
						Aliases:  []string{"atto evmos"},
					},
					{
						Denom:    "evmos",
						Exponent: 18,
					},
				},
			},
			banktypes.Metadata{
				Base:        "aevmos",
				Display:     "evmos",
				Name:        "Evmos",
				Symbol:      "EVMOS",
				Description: "EVM, staking and governance denom of Evmos",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    "aevmos",
						Exponent: 0,
						Aliases:  []string{"atto evmos"},
					},
				},
			},
			true,
		},
		{
			"different denom units",
			banktypes.Metadata{
				Base:        "aevmos",
				Display:     "evmos",
				Name:        "Evmos",
				Symbol:      "EVMOS",
				Description: "EVM, staking and governance denom of Evmos",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    "aevmos",
						Exponent: 0,
						Aliases:  []string{"atto evmos"},
					},
					{
						Denom:    "uevmos",
						Exponent: 12,
						Aliases:  []string{"micro evmos"},
					},
					{
						Denom:    "evmos",
						Exponent: 18,
					},
				},
			},
			banktypes.Metadata{
				Base:        "aevmos",
				Display:     "evmos",
				Name:        "Evmos",
				Symbol:      "EVMOS",
				Description: "EVM, staking and governance denom of Evmos",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    "aevmos",
						Exponent: 0,
						Aliases:  []string{"atto evmos"},
					},
					{
						Denom:    "Uevmos",
						Exponent: 12,
						Aliases:  []string{"micro evmos"},
					},
					{
						Denom:    "evmos",
						Exponent: 18,
					},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := EqualMetadata(tc.metadataA, tc.metadataB)
		if tc.expError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestEqualAliases(t *testing.T) {
	testCases := []struct {
		name     string
		aliasesA []string
		aliasesB []string
		expEqual bool
	}{
		{
			"empty",
			[]string{},
			[]string{},
			true,
		},
		{
			"different lengths",
			[]string{},
			[]string{"atto evmos"},
			false,
		},
		{
			"different values",
			[]string{"attoevmos"},
			[]string{"atto evmos"},
			false,
		},
		{
			"same values, unsorted",
			[]string{"atto evmos", "aevmos"},
			[]string{"aevmos", "atto evmos"},
			false,
		},
		{
			"same values, sorted",
			[]string{"aevmos", "atto evmos"},
			[]string{"aevmos", "atto evmos"},
			true,
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.expEqual, EqualStringSlice(tc.aliasesA, tc.aliasesB), tc.name)
	}
}
