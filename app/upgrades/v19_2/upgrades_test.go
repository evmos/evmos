package v192_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	v192 "github.com/evmos/evmos/v20/app/upgrades/v19_2"
	testnetwork "github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	"github.com/evmos/evmos/v20/x/evm/statedb"
)

const expNonce uint64 = 3

var (
	expCodeHash   = common.HexToHash("0x7b477c761b4d0469f03f27ba58d0a7eacbfdd62b69b82c6c683ae5f81c67fe80")
	otherCodeHash = crypto.Keccak256([]byte("random"))
	expBalance    = big.NewInt(10)
)

var tokenPairsSeed = []erc20types.TokenPair{
	{
		Erc20Address:  "0xb72A7567847abA28A2819B855D7fE679D4f59846",
		Denom:         "erc20/0xb72A7567847abA28A2819B855D7fE679D4f59846",
		Enabled:       true,
		ContractOwner: erc20types.OWNER_EXTERNAL,
	},
	{
		Erc20Address:  "0xc03345448969Dd8C00e9E4A85d2d9722d093aF8E",
		Denom:         "ibc/6B3FCE336C3465D3B72F7EFB4EB92FC521BC480FE9653F627A0BD0237DF213F3",
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE,
	},
	{
		Erc20Address:  "0xb98e169C37ce30Dd47Fdad1f9726Fb832191e60b",
		Denom:         "erc20/0xb98e169C37ce30Dd47Fdad1f9726Fb832191e60b",
		Enabled:       true,
		ContractOwner: erc20types.OWNER_EXTERNAL,
	},
	{
		Erc20Address:  "0x5db67696C3c088DfBf588d3dd849f44266ff0ffa",
		Denom:         "ibc/AB40F54CC4BAF9C4FDB5CB5BDFC4D6CDB475BFE84BCEAF4F72F02EDAEED60F05",
		Enabled:       false,
		ContractOwner: erc20types.OWNER_MODULE,
	},
	{
		Erc20Address:  "0xc76A204AEA61a68a3B1f97B8E70286CD42B020D2",
		Denom:         "ibc/26E6508A1757E12B15A087E951F5D35E73CF036F0D97BC809E1598D1DD870BED",
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE,
	},
}

func TestAddCodeToERC20Extensions(t *testing.T) {
	var (
		network *testnetwork.UnitTestNetwork
		// address of an erc20 contract from a native ERC20
		ibcCoinAddr = common.HexToAddress(tokenPairsSeed[1].Erc20Address)
		// address of an erc20 contract from a IBC coin
		er20Addr = common.HexToAddress(tokenPairsSeed[0].Erc20Address)
	)
	testCases := []struct {
		name      string
		malleate  func(ctx sdk.Context)
		postCheck func(t *testing.T, ctx sdk.Context, tp erc20types.TokenPair)
	}{
		{
			name:     "all non-existent accounts",
			malleate: func(sdk.Context) {},
			postCheck: func(t *testing.T, ctx sdk.Context, p erc20types.TokenPair) {
				contractAddr := common.HexToAddress(p.Erc20Address)
				acc := network.App.AccountKeeper.GetAccount(ctx, contractAddr.Bytes())
				// no changes should be applied to native erc20s
				if p.IsNativeERC20() {
					require.Nil(t, acc)
					return
				}
				ethAddr := common.BytesToAddress(acc.GetAddress().Bytes())
				codeHash := network.App.EvmKeeper.GetCodeHash(ctx, ethAddr)
				require.Equal(t, codeHash.String(), expCodeHash.String())
			},
		},
		{
			name: "all existent accounts",
			malleate: func(ctx sdk.Context) {
				// set existent account to native ERC20
				err := network.App.EvmKeeper.SetAccount(ctx, er20Addr, statedb.Account{
					Nonce:    expNonce,
					Balance:  expBalance,
					CodeHash: otherCodeHash,
				})
				require.NoError(t, err)
				// set existent account to IBC coin
				err = network.App.EvmKeeper.SetAccount(ctx, ibcCoinAddr, statedb.Account{
					Nonce:    expNonce,
					Balance:  expBalance,
					CodeHash: otherCodeHash,
				})
				require.NoError(t, err)
			},
			postCheck: func(t *testing.T, ctx sdk.Context, p erc20types.TokenPair) {
				addr := common.HexToAddress(p.Erc20Address)
				acct := network.App.EvmKeeper.GetAccount(ctx, addr)
				switch common.HexToAddress(p.Erc20Address) {
				case er20Addr:
					require.NotNil(t, acct)
					require.Equal(t, acct.Balance, expBalance)
					require.Equal(t, acct.Nonce, expNonce)
					require.Equal(t, acct.CodeHash, otherCodeHash)
				case ibcCoinAddr:
					require.NotNil(t, acct)
					require.Equal(t, acct.Balance, expBalance)
					require.Equal(t, acct.Nonce, expNonce)
					// only code hash should be updated
					require.Equal(t, acct.CodeHash, expCodeHash.Bytes())
				default:
					if p.IsNativeERC20() {
						require.Nil(t, acct)
						return
					}
					require.NotNil(t, acct)
					require.Equal(t, acct.CodeHash, expCodeHash.Bytes())
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			erc20GenesisState := erc20types.DefaultGenesisState()
			erc20GenesisState.TokenPairs = []erc20types.TokenPair{}
			genesis := testnetwork.CustomGenesisState{
				erc20types.ModuleName: erc20GenesisState,
			}
			network = testnetwork.NewUnitTestNetwork(
				testnetwork.WithCustomGenesis(genesis))

			ctx := network.GetContext()

			// seed the token pairs
			for _, p := range tokenPairsSeed {
				network.App.Erc20Keeper.SetToken(ctx, p)
			}

			tc.malleate(ctx)

			logger := ctx.Logger()
			err := v192.AddCodeToERC20Extensions(ctx, logger, network.App.Erc20Keeper)
			require.NoError(t, err)

			code := network.App.EvmKeeper.GetCode(ctx, expCodeHash)
			require.True(t, len(code) > 0)

			pairs := network.App.Erc20Keeper.GetTokenPairs(ctx)
			require.Equal(t, len(tokenPairsSeed), len(pairs))
			for _, p := range pairs {
				tc.postCheck(t, ctx, p)
			}
		})
	}
}
