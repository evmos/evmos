package bank_test

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/bank"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/stretchr/testify/require"
)

func TestParseBalancesArgs(t *testing.T) {
	testCases := []struct {
		name    string
		args    []interface{}
		expAddr sdk.AccAddress
		expErr  string
	}{
		{
			"invalid length",
			nil,
			nil,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			"invalid type",
			[]interface{}{"address"},
			nil,
			"invalid account address",
		},
		{
			"success",
			[]interface{}{common.Address{}},
			common.Address{}.Bytes(),
			"",
		},
	}

	for _, tc := range testCases {
		address, err := bank.ParseBalancesArgs(tc.args)
		if tc.expErr != "" {
			require.True(t, strings.Contains(err.Error(), tc.expErr), err.Error())
		} else {
			require.NoError(t, err)
			require.Equal(t, tc.expAddr, address)
		}
	}
}
