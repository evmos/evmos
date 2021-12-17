package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/incentives/types"
)

// GetIncentivesGasMeters - get all registered GasMeters per Incentive
func (k Keeper) GetIncentivesGasMeters(ctx sdk.Context) []types.GasMeter {
	gms := []types.GasMeter{}

	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixGasMeter)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		contract, userAddress := types.SplitGasMeterKey(iterator.Key())
		gas := sdk.BigEndianToUint64(iterator.Value())

		gm := types.Gasgm{
			Contract:       contract.Hex(),
			Participant:    userAddress.Hex(),
			CummulativeGas: gas,
		}

		gms = append(gms, gm)
	}

	return gms
}

// GetGasMetersByContract - get all registered GasMeters per contract
func (k Keeper) GetGasMetersByContract(
	ctx sdk.Context,
	contract common.Address,
) []types.GasMeter {
	gms := []types.GasMeter{}
	store := ctx.KVStore(k.storeKey)
	key := append(types.KeyPrefixGasMeter, contract.Bytes()...)

	iterator := sdk.KVStorePrefixIterator(store, key)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		contract, userAddress := types.SplitGasMeterKey(iterator.Key())
		gas := sdk.BigEndianToUint64(iterator.Value())
		gm := types.GasMeter{
			Contract:       contract.Hex(),
			Participant:    userAddress.Hex(),
			CummulativeGas: gas,
		}

		gms = append(gms, gm)
	}

	return gms
}

func (k Keeper) IterateIncentiveGasMeters(
	ctx sdk.Context,
	contract common.Address,
	handlerFn func(gm types.GasMeter) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.KeyPrefixGasMeter, contract.Bytes()...)

	iterator := sdk.KVStorePrefixIterator(store, key)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		contract, userAddress := types.SplitGasMeterKey(iterator.Key())
		gas := sdk.BigEndianToUint64(iterator.Value())

		gm := types.GasMeter{
			Contract:       contract.Hex(),
			Participant:    userAddress.Hex(),
			CummulativeGas: gas,
		}

		if handlerFn(gm) {
			break
		}
	}
}

// GetIncentiveGasMeter - get cumulativeGas from participant
func (k Keeper) GetIncentiveGasMeter(
	ctx sdk.Context,
	contract, userAddress common.Address,
) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), append(types.KeyPrefixGasMeter, contract.Bytes()...))

	bz := store.Get(userAddress.Bytes())
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetGasMeter stores a gasMeter
func (k Keeper) SetGasMeter(ctx sdk.Context, gm types.GasMeter) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixGasMeter)


	key := append(types.KeyPrefixGasMeter, []byte(gm.Contract), gm.  )

	key := common.HexToAddress(incentive.Contract)
	bz := k.cdc.MustMarshal(&incentive)
	store.Set(key.Bytes(), bz)
}

// DeleteIncentive removes a token pair.
func (k Keeper) DeleteGasMeter(ctx sdk.Context, gn types.GasMeter) {
	// TODO
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIncentive)
	key := common.HexToAddress(incentive.Contract)
	store.Delete(key.Bytes())
}
