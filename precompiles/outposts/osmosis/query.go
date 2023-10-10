package osmosis

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/evmos/evmos/v14/precompiles/authorization"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
	"github.com/evmos/evmos/v14/precompiles/ics20"
)

// Allowance returns the remaining allowance of for a combination of grantee - granter.
// The grantee is the smart contract that was authorized by the granter to spend.
func (p Precompile) Allowance(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// append here the msg type. Will always be the TransferMsg
	// for this precompile
	args = append(args, ics20.TransferMsgURL)

	grantee, granter, msg, err := authorization.CheckAllowanceArgs(args)
	if err != nil {
		return nil, err
	}

	msgAuthz, _ := p.AuthzKeeper.GetAuthorization(ctx, grantee.Bytes(), granter.Bytes(), msg)

	if msgAuthz == nil {
		// return empty array
		return method.Outputs.Pack([]ics20.Allocation{})
	}

	transferAuthz, ok := msgAuthz.(*transfertypes.TransferAuthorization)
	if !ok {
		return nil, fmt.Errorf(cmn.ErrInvalidType, "transfer authorization", &transfertypes.TransferAuthorization{}, transferAuthz)
	}

	// need to convert to ics20.Allocation (uses big.Int)
	// because ibc Allocation has sdkmath.Int
	allocs := make([]ics20.Allocation, len(transferAuthz.Allocations))
	for i, a := range transferAuthz.Allocations {
		spendLimit := make([]cmn.Coin, len(a.SpendLimit))
		for j, c := range a.SpendLimit {
			spendLimit[j] = cmn.Coin{
				Denom:  c.Denom,
				Amount: c.Amount.BigInt(),
			}
		}

		allocs[i] = ics20.Allocation{
			SourcePort:    a.SourcePort,
			SourceChannel: a.SourceChannel,
			SpendLimit:    spendLimit,
			AllowList:     a.AllowList,
		}
	}

	return method.Outputs.Pack(allocs)
}
