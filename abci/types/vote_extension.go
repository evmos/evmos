package types

import (
	"fmt"

	"github.com/evmos/evmos/v16/types"
)

func (ve EVMVoteExtension) Validate() error {
	if ve.Height < 0 {
		return fmt.Errorf("height cannot be negative: %d", ve.Height)
	}

	if ve.BaseFee.IsNegative() {
		return fmt.Errorf("base fee cannot be negative: %s", ve.BaseFee)
	}

	if ve.BlockGasUsed == 0 {
		return fmt.Errorf("block gas used cannot be zero")
	}

	if err := types.ValidateAddress(ve.Miner); err != nil {
		return fmt.Errorf("miner address is invalid: %w", err)
	}

	if len(ve.ExtraData) > 32 {
		return fmt.Errorf("extra data cannot exceed 32 bytes")
	}

	return nil
}
