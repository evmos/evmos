package types_test

import (
	"testing"

	"github.com/evmos/evmos/v20/x/erc20/types"
	"github.com/stretchr/testify/require"
)

func TestNewERC20Data(t *testing.T) {
	data := types.NewERC20Data("test", "ERC20", uint8(18))
	exp := types.ERC20Data{Name: "test", Symbol: "ERC20", Decimals: 0x12}
	require.Equal(t, exp, data)
}
