// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package upgrade

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// GetBalance returns the account balance for the given address bech32 string through
// the CLI query.
func (m *Manager) GetBalance(ctx context.Context, chainID, address string) (sdk.Coins, error) {
	queryArgs := QueryArgs{
		Module:     "bank",
		SubCommand: "balances",
		Args:       []string{address},
		ChainID:    chainID,
	}

	exec, err := m.CreateModuleQueryExec(queryArgs)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("create exec error: %w", err)
	}

	outBuff, errBuff, err := m.RunExec(ctx, exec)
	if err != nil {
		return sdk.Coins{}, fmt.Errorf("run exec error: %w", err)
	}
	if errBuff.String() != "" {
		return sdk.Coins{}, fmt.Errorf("evmos query error: %s", errBuff.String())
	}

	return UnpackBalancesResponse(m.ProtoCodec, outBuff.String())
}

// UnpackBalancesResponse unpacks the balances response from the given output of the bank balances query.
func UnpackBalancesResponse(cdc *codec.ProtoCodec, out string) (sdk.Coins, error) {
	var balances banktypes.QueryAllBalancesResponse
	if err := cdc.UnmarshalJSON([]byte(out), &balances); err != nil {
		return sdk.Coins{}, fmt.Errorf("failed to unmarshal balances: %w", err)
	}

	return balances.Balances, nil
}
