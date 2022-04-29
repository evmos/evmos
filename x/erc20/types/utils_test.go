package types

import (
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

func TestSanitizeERC20Name(t *testing.T) {
	testCases := []struct {
		name         string
		erc20Name    string
		expErc20Name string
	}{
		{"name contains ' Token'", "Lucky Token", "lucky"},
		{"name contains ' Coin'", "Otter Coin", "otter"},
		{"name contains ' Token' and ' Coin'", "Lucky Token Coin", "lucky"},
		{"multiple words", "Hextris Early Access Demo", "hextris_early_access_demo"},
		{"single word name: Token", "Token", "token"},
		{"single word name: Coin", "Coin", "coin"},
	}

	for _, tc := range testCases {
		name := SanitizeERC20Name(tc.erc20Name)
		require.Equal(t, tc.expErc20Name, name, tc.name)
	}
}

func TestEqualMetadata(t *testing.T) {

	base := "CoinBase"
	display := "CoinDisplay"
	name := "CoinName"
	symbol := "CoinATOM"
	description := "ATOM Coin"
	decimal := uint32(18)

	// for metadata_A
	denomUnits_A := []*banktypes.DenomUnit{
		{
			Denom:    base,
			Exponent: 0,
			Aliases:  []string{base, "moreInfo"},
		},
		{
			Denom:    display,
			Exponent: decimal,
			Aliases:  []string{display, "moreInfo"},
		}}
	metadata_A := banktypes.Metadata{
		Base:        base,
		Display:     display,
		Name:        name,
		Symbol:      symbol,
		Description: description,
		DenomUnits:  denomUnits_A,
	}

	// for metadata_B
	denomUnits_B := []*banktypes.DenomUnit{
		{
			Denom:    base,
			Exponent: 0,
			Aliases:  []string{"moreInfo", base},
		},
		{
			Denom:    display,
			Exponent: decimal,
			Aliases:  []string{display, "moreInfo"},
		}}
	metadata_B := banktypes.Metadata{
		Base:        base,
		Display:     display,
		Name:        name,
		Symbol:      symbol,
		Description: description,
		DenomUnits:  denomUnits_B,
	}

	// for metadata_C
	denomUnits_C := []*banktypes.DenomUnit{
		{
			Denom:    base,
			Exponent: 0,
			Aliases:  []string{"moreInfo_YES", base},
		},
		{
			Denom:    display,
			Exponent: decimal,
			Aliases:  []string{display, "moreInfo"},
		}}
	metadata_C := banktypes.Metadata{
		Base:        base,
		Display:     display,
		Name:        name,
		Symbol:      symbol,
		Description: description,
		DenomUnits:  denomUnits_C,
	}

	// metadata list
	metadataList := []*banktypes.Metadata{
		&metadata_A,
		&metadata_B,
		&metadata_C,
	}

	// validate each metadata
	for _, md := range metadataList {
		require.NoError(t, md.Validate())
	}

	require.NoError(t, EqualMetadata(metadata_A, metadata_B))
	require.NotEqual(t, EqualMetadata(metadata_A, metadata_C), nil)
}
