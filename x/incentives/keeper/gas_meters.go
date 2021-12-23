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

		gm := types.GasMeter{
			Contract:      contract.String(),
			Participant:   userAddress.String(),
			CumulativeGas: gas,
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

	k.IterateIncentiveGasMeters(
		ctx, contract,
		func(gm types.GasMeter) (stop bool) {
			gms = append(gms, gm)
			return false
		})

	return gms
}

func (k Keeper) IterateIncentiveGasMeters(
	ctx sdk.Context,
	contract common.Address,
	handlerFn func(gm types.GasMeter) (stop bool),
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixGasMeter)

	iterator := sdk.KVStorePrefixIterator(store, contract.Bytes())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		contract, userAddress := types.SplitGasMeterKey(iterator.Key())
		gas := sdk.BigEndianToUint64(iterator.Value())

		gm := types.GasMeter{
			Contract:      contract.String(),
			Participant:   userAddress.String(),
			CumulativeGas: gas,
		}

		if handlerFn(gm) {
			break
		}
	}
}

// GetIncentiveGasMeter - get cumulativeGas from participant
func (k Keeper) GetIncentiveGasMeter(
	ctx sdk.Context,
	contract, participant common.Address,
) (uint64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixGasMeter)
	key := append(contract.Bytes(), participant.Bytes()...)

	bz := store.Get(key)
	if len(bz) == 0 {
		return 0, false
	}

	return sdk.BigEndianToUint64(bz), true
}

// SetGasMeter stores a gasMeter
func (k Keeper) SetGasMeter(ctx sdk.Context, gm types.GasMeter) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixGasMeter)
	contract := common.HexToAddress(gm.Contract)
	participant := common.HexToAddress(gm.Participant)
	key := append(contract.Bytes(), participant.Bytes()...)
	store.Set(key, sdk.Uint64ToBigEndian(gm.CumulativeGas))
}

// DeleteIncentive removes a token pair.
func (k Keeper) DeleteGasMeter(ctx sdk.Context, gm types.GasMeter) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixGasMeter)
	contract := common.HexToAddress(gm.Contract)
	participant := common.HexToAddress(gm.Participant)
	key := append(contract.Bytes(), participant.Bytes()...)
	store.Delete(key)
}
