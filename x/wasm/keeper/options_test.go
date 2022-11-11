package keeper

import (
	"reflect"
	"testing"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/evoblockchain/evoblock/x/wasm/keeper/wasmtesting"
	"github.com/evoblockchain/evoblock/x/wasm/types"
)

func TestConstructorOptions(t *testing.T) {
	specs := map[string]struct {
		srcOpt Option
		verify func(*testing.T, Keeper)
	}{
		"wasm engine": {
			srcOpt: WithWasmEngine(&wasmtesting.MockWasmer{}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockWasmer{}, k.wasmVM)
			},
		},
		"message handler": {
			srcOpt: WithMessageHandler(&wasmtesting.MockMessageHandler{}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockMessageHandler{}, k.messenger)
			},
		},
		"query plugins": {
			srcOpt: WithQueryHandler(&wasmtesting.MockQueryHandler{}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockQueryHandler{}, k.wasmVMQueryHandler)
			},
		},
		"message handler decorator": {
			srcOpt: WithMessageHandlerDecorator(func(old Messenger) Messenger {
				require.IsType(t, &MessageHandlerChain{}, old)
				return &wasmtesting.MockMessageHandler{}
			}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockMessageHandler{}, k.messenger)
			},
		},
		"query plugins decorator": {
			srcOpt: WithQueryHandlerDecorator(func(old WasmVMQueryHandler) WasmVMQueryHandler {
				require.IsType(t, QueryPlugins{}, old)
				return &wasmtesting.MockQueryHandler{}
			}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockQueryHandler{}, k.wasmVMQueryHandler)
			},
		},
		"coin transferrer": {
			srcOpt: WithCoinTransferrer(&wasmtesting.MockCoinTransferrer{}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockCoinTransferrer{}, k.bank)
			},
		},
		"costs": {
			srcOpt: WithGasRegister(&wasmtesting.MockGasRegister{}),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, &wasmtesting.MockGasRegister{}, k.gasRegister)
			},
		},
		"api costs": {
			srcOpt: WithAPICosts(1, 2),
			verify: func(t *testing.T, k Keeper) {
				t.Cleanup(setApiDefaults)
				assert.Equal(t, uint64(1), costHumanize)
				assert.Equal(t, uint64(2), costCanonical)
			},
		},
		"max recursion query limit": {
			srcOpt: WithMaxQueryStackSize(1),
			verify: func(t *testing.T, k Keeper) {
				assert.IsType(t, uint32(1), k.maxQueryStackSize)
			},
		},
		"accepted account types": {
			srcOpt: WithAcceptedAccountTypesOnContractInstantiation(&authtypes.BaseAccount{}, &vestingtypes.ContinuousVestingAccount{}),
			verify: func(t *testing.T, k Keeper) {
				exp := map[reflect.Type]struct{}{
					reflect.TypeOf(&authtypes.BaseAccount{}):                 {},
					reflect.TypeOf(&vestingtypes.ContinuousVestingAccount{}): {},
				}
				assert.Equal(t, exp, k.acceptedAccountTypes)
			},
		},
		"account pruner": {
			srcOpt: WithAccountPruner(VestingCoinBurner{}),
			verify: func(t *testing.T, k Keeper) {
				assert.Equal(t, VestingCoinBurner{}, k.accountPruner)
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			k := NewKeeper(nil, nil, paramtypes.NewSubspace(nil, nil, nil, nil, ""), authkeeper.AccountKeeper{}, &bankkeeper.BaseKeeper{}, stakingkeeper.Keeper{}, distributionkeeper.Keeper{}, nil, nil, nil, nil, nil, nil, "tempDir", types.DefaultWasmConfig(), AvailableCapabilities, spec.srcOpt)
			spec.verify(t, k)
		})
	}
}

func setApiDefaults() {
	costHumanize = DefaultGasCostHumanAddress * DefaultGasMultiplier
	costCanonical = DefaultGasCostCanonicalAddress * DefaultGasMultiplier
}
