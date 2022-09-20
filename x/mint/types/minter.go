package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMinter returns a new Minter object with the given block
// provisions values.
func NewMinter(dailyProvisions sdk.Dec, lastMintTime int64) Minter {
	return Minter{
		LastMintTime:    lastMintTime,
		DailyProvisions: dailyProvisions,
	}
}

// InitialMinter returns an initial Minter object.
func InitialMinter() Minter {
	return NewMinter(sdk.NewDec(0), 0)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain.
func DefaultInitialMinter() Minter {
	return InitialMinter()
}

// Validate validates minter. Returns nil on success, error otherewise.
func (m Minter) Validate() error {
	return nil
}

// BlockProvision returns the provisions for a block based on the block
// provisions rate.
func (m Minter) BlockProvision(time int64, params Params) sdk.Coin {
	provisionAmt := m.DailyProvisions.Mul(sdk.NewDec(time - m.LastMintTime)).Quo(sdk.NewDec(86400))
	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}
