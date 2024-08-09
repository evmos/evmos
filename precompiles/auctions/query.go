// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package auctions

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	auctionstypes "github.com/evmos/evmos/v19/x/auctions/types"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
)

// AuctionInfoMethod is the method name for the auction info query
const AuctionInfoMethod = "auctionInfo"

// AuctionInfo defines the query for information about the current auction.
func (p Precompile) AuctionInfo(
	ctx sdk.Context,
	_ *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf(ErrInvalidInputLength, len(args))
	}

	res, err := p.auctionsKeeper.AuctionInfo(ctx, &auctionstypes.QueryCurrentAuctionInfoRequest{})
	if err != nil {
		return nil, err
	}

	out := new(AuctionInfoOutput).FromResponse(res)

	return out.Pack(method.Outputs)
}
