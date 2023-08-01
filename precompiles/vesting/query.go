package vesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// BalancesMethod defines the ABI method name for the Balances query.
	BalancesMethod = "balances"
)

// Balances queries the balances of a clawback vesting account.
func (p Precompile) Balances(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, err := NewBalancesRequest(args)
	if err != nil {
		return nil, err
	}

	response, err := p.vestingKeeper.Balances(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	out := new(BalancesOutput).FromResponse(response)

	return method.Outputs.Pack(out.Locked, out.Unvested, out.Vested)
}
