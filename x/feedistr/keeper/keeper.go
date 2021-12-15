package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/evmos/x/feedistr/types"
)

// Keeper of this module maintains .
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace
}

// NewKeeper creates new instances of the distribution Keeper
func NewKeeper(
	storeKey sdk.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		paramstore: ps,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetContractWithdrawAddresses - get all registered contracts with their corresponding
// withdraw address
func (k Keeper) GetContractWithdrawAddresses(ctx sdk.Context) []types.ContractWithdrawAddress {
	withdrawalAddresses := []types.ContractWithdrawAddress{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixContractOwner)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		contractWithdrawAddress := types.ContractWithdrawAddress{
			ContractAddress: common.BytesToAddress(iterator.Key()).Hex(),
			WithdrawAddress: common.BytesToAddress(iterator.Value()).Hex(),
		}
		withdrawalAddresses = append(withdrawalAddresses, contractWithdrawAddress)
	}

	return withdrawalAddresses
}

func (k Keeper) GetContractWithdrawAddress(ctx sdk.Context, contract common.Address) (common.Address, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixContractOwner)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return common.Address{}, false
	}

	return common.BytesToAddress(bz), true
}

func (k Keeper) HasContractWithdrawAddress(ctx sdk.Context, contract common.Address) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixContractOwner)
	return store.Has(contract.Bytes())
}

// SetContractWithdrawAddress
func (k Keeper) SetContractWithdrawAddress(ctx sdk.Context, contractAddr, withdrawAddr common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixContractOwner)
	store.Set(contractAddr.Bytes(), withdrawAddr.Bytes())
}

// DeleteContractWithdrawAddress removes a contract withdraw address.
func (k Keeper) DeleteContractWithdrawAddress(ctx sdk.Context, contract common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixContractOwner)
	store.Delete(contract.Bytes())
}

func (k Keeper) HasContractWithdrawAddressInverse(ctx sdk.Context, withdrawAddr, contractAddr common.Address) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), append(types.KeyPrefixContractOwnerInverse, withdrawAddr.Bytes()...))
	return store.Has(contractAddr.Bytes())
}

// SetContractWithdrawAddress
func (k Keeper) SetContractWithdrawAddressInverse(ctx sdk.Context, contractAddr, withdrawAddr common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), append(types.KeyPrefixContractOwnerInverse, withdrawAddr.Bytes()...))
	store.Set(contractAddr.Bytes(), []byte{0x1})
}

// DeleteContractWithdrawAddressInverse removes a contract from the withdraw address records.
func (k Keeper) DeleteContractWithdrawAddressInverse(ctx sdk.Context, withdrawAddr, contractAddr common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), append(types.KeyPrefixContractOwnerInverse, withdrawAddr.Bytes()...))
	store.Delete(contractAddr.Bytes())
}
