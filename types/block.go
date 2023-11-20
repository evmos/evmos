// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	math "math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BlockGasLimit returns the max gas (limit) defined in the block gas meter. If the meter is not
// set, it returns the max gas from the application consensus params.
// NOTE: see https://github.com/cosmos/cosmos-sdk/issues/9514 for full reference
func BlockGasLimit(ctx sdk.Context) uint64 {
	blockGasMeter := ctx.BlockGasMeter()

	// Get the limit from the gas meter only if its not null and not an InfiniteGasMeter
	if blockGasMeter != nil && blockGasMeter.Limit() != 0 {
		return blockGasMeter.Limit()
	}

	// Otherwise get from the consensus parameters
	cp := ctx.ConsensusParams()
	if cp.Block == nil {
		return 0
	}

	maxGas := cp.Block.MaxGas

	// Setting max_gas to -1 in Tendermint means there is no limit on the maximum gas consumption for transactions
	// https://github.com/cometbft/cometbft/blob/v0.37.2/proto/tendermint/types/params.proto#L25-L27
	if maxGas == -1 {
		return math.MaxUint64
	}

	if maxGas > 0 {
		return uint64(maxGas) // #nosec G701 -- maxGas is int64 type. It can never be greater than math.MaxUint64
	}

	return 0
}
