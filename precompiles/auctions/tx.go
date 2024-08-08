package auctions

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	precompilecommon "github.com/evmos/evmos/v19/precompiles/common"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
)

const (
	// BidMethod defines the ABI method name for the auctions
	// Bid transaction.
	BidMethod = "setWithdrawAddress"
	// DepositCoinMethod defines the ABI method name for the auctions
	// DepositCoin transaction.
	DepositCoinMethod = "withdrawDelegatorRewards"
)

// Bid bids on the current auction with a specified Evmos amount that must be higher than the highest bid.
func (p *Precompile) Bid(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {

	sender, msgBid := NewMsgBid(args)

	_, err := p.auctionsKeeper.Bid(ctx, msgBid)
	if err != nil {
		return nil, err
	}

	// emits an event for the Bid transaction.
	if err := p.EmitBidEvent(ctx, stateDB, sender, msgBid.Amount.Amount.BigInt()); err != nil {
		return nil, err
	}

	return precompilecommon.TrueValue, nil
}

// DepositCoin deposits coins into the auction collector module to be used in the following auction.
func (p *Precompile) DepositCoin(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {

	sender, msgDepositCoin := NewMsgDepositCoin(args)

	_, err := p.auctionsKeeper.DepositCoin(ctx, msgDepositCoin)
	if err != nil {
		return nil, err
	}

	// emits an event for the DepositCoin transaction.
	if err := p.EmitDepositCoinEvent(ctx, stateDB, sender, msgDepositCoin.Amount.Denom, msgDepositCoin.Amount.Amount.BigInt()); err != nil {
		return nil, err
	}

	return precompilecommon.TrueValue, nil
}
