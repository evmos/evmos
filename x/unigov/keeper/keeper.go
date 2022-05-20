package keeper

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/Canto-Network/canto/v3/x/unigov/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		storeKey   sdk.StoreKey
		cdc        codec.BinaryCodec
		paramstore paramtypes.Subspace

		mapContractAddr common.Address
		accKeeper       types.AccountKeeper
		erc20Keeper     types.ERC20Keeper
	}
)

func NewKeeper(
	storeKey sdk.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,

	mca common.Address,
	ak types.AccountKeeper,
	ek types.ERC20Keeper,

) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	mca = common.HexToAddress("0000000000000000000000000000000000000000")

	return Keeper{

		cdc:             cdc,
		storeKey:        storeKey,
		mapContractAddr: mca,
		paramstore:      ps,
		accKeeper:       ak,
		erc20Keeper:     ek,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
