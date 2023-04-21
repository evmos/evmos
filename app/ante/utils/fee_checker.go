package utils

import sdk "github.com/cosmos/cosmos-sdk/types"

// TxFeeChecker check if the provided fee is enough and returns the effective fee and tx priority,
// the effective fee should be deducted later, and the priority should be returned in abci response.
type TxFeeChecker func(ctx sdk.Context, feeTx sdk.FeeTx) (sdk.Coins, int64, error)
