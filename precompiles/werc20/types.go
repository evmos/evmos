package werc20

import (
	"fmt"
	"math/big"
)

func ParseWithdrawArgs(args []interface{}) (
	amount *big.Int, err error,
) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments; expected 2; got: %d", len(args))
	}

	amount, ok := args[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %v", args[1])
	}

	return amount, nil
}
