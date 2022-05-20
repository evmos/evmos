package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/Canto-Network/canto/v3/x/unigov/types"
	"github.com/ethereum/go-ethereum/"
)

type (
	Keeper struct {
		
		cdc      	codec.BinaryCodec
		storeKey 	sdk.StoreKey
		memKey   	sdk.StoreKey
		paramstore	paramtypes.Subspace

		mapContractAddr common.Address
		accKeeper   types.AccountKeeper
		erc20Keeper types.ERC20Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,

	
	
	addr common.Address
	ak   types.AccountKeeper,
	ek types.ERC20Keeper,
	
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		
		cdc:      	cdc,
		storeKey: 	storeKey,
		memKey:   	memKey,
		mapContractAddr: addr,
		paramstore:	ps,
		accKeeper:      ak,
		erc20Keeper:    ek,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
