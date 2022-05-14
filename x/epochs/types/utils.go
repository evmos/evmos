package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DurationToBz parses time duration to maintain number-compatible ordering
func DurationToBz(duration time.Duration) []byte {
	return sdk.Uint64ToBigEndian(uint64(duration.Milliseconds()))
}
