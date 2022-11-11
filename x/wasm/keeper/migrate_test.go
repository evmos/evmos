package keeper

import (
	"bytes"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/require"

	"github.com/evoblockchain/evoblock/x/wasm/types"
)

func TestMigrate1To2(t *testing.T) {
	ctx, keepers := CreateTestInput(t, false, AvailableCapabilities)
	wasmKeeper := keepers.WasmKeeper

	deposit := sdk.NewCoins(sdk.NewInt64Coin("denom", 100000))
	creator := sdk.AccAddress(bytes.Repeat([]byte{1}, address.Len))
	keepers.Faucet.Fund(ctx, creator, deposit...)
	example := StoreHackatomExampleContract(t, ctx, keepers)

	initMsg := HackatomExampleInitMsg{
		Verifier:    RandomAccountAddress(t),
		Beneficiary: RandomAccountAddress(t),
	}
	initMsgBz, err := json.Marshal(initMsg)
	require.NoError(t, err)

	em := sdk.NewEventManager()

	// create with no balance is also legal
	gotContractAddr1, _, err := keepers.ContractKeeper.Instantiate(ctx.WithEventManager(em), example.CodeID, creator, nil, initMsgBz, "demo contract 1", nil)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	gotContractAddr2, _, err := keepers.ContractKeeper.Instantiate(ctx.WithEventManager(em), example.CodeID, creator, nil, initMsgBz, "demo contract 1", nil)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	gotContractAddr3, _, err := keepers.ContractKeeper.Instantiate(ctx.WithEventManager(em), example.CodeID, creator, nil, initMsgBz, "demo contract 1", nil)

	info1 := wasmKeeper.GetContractInfo(ctx, gotContractAddr1)
	info2 := wasmKeeper.GetContractInfo(ctx, gotContractAddr2)
	info3 := wasmKeeper.GetContractInfo(ctx, gotContractAddr3)

	// remove key
	ctx.KVStore(wasmKeeper.storeKey).Delete(types.GetContractByCreatorSecondaryIndexKey(creator, info1.Created.Bytes(), gotContractAddr1))
	ctx.KVStore(wasmKeeper.storeKey).Delete(types.GetContractByCreatorSecondaryIndexKey(creator, info2.Created.Bytes(), gotContractAddr2))
	ctx.KVStore(wasmKeeper.storeKey).Delete(types.GetContractByCreatorSecondaryIndexKey(creator, info3.Created.Bytes(), gotContractAddr3))

	// migrator
	migrator := NewMigrator(*wasmKeeper)
	migrator.Migrate1to2(ctx)

	// check new store
	var allContract []string
	wasmKeeper.IterateContractsByCreator(ctx, creator, func(addr sdk.AccAddress) bool {
		allContract = append(allContract, addr.String())
		return false
	})

	require.Equal(t, []string{gotContractAddr1.String(), gotContractAddr2.String(), gotContractAddr3.String()}, allContract)
}
