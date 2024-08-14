package v192_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	v192 "github.com/evmos/evmos/v19/app/upgrades/v19_2"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	testnetwork "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v19/types"
	erc20types "github.com/evmos/evmos/v19/x/erc20/types"
	"github.com/stretchr/testify/require"
)

const expCodeHash = "0x7b477c761b4d0469f03f27ba58d0a7eacbfdd62b69b82c6c683ae5f81c67fe80"

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
	// initialize network without wevmos
	erc20GenesisState := erc20types.DefaultGenesisState()
	erc20GenesisState.TokenPairs = []erc20types.TokenPair{}
	genesis := network.CustomGenesisState{
		erc20types.ModuleName: erc20GenesisState,
	}
	network := testnetwork.NewUnitTestNetwork(
		testnetwork.WithCustomGenesis(genesis))
	ctx := network.GetContext()

	// check code does not exist
	code := network.App.EvmKeeper.GetCode(ctx, common.HexToHash(expCodeHash))
	require.Len(t, code, 0)

	// seed the token pairs
	for _, p := range tokenPairsSeed {
		network.App.Erc20Keeper.SetToken(ctx, p)
	}

	logger := ctx.Logger()
	err := v192.AddCodeToERC20Extensions(ctx, logger, network.App.Erc20Keeper, network.App.EvmKeeper)
	require.NoError(t, err)

	code = network.App.EvmKeeper.GetCode(ctx, common.HexToHash(expCodeHash))
	require.True(t, len(code) > 0)

	pairs := network.App.Erc20Keeper.GetTokenPairs(ctx)
	require.Equal(t, len(tokenPairsSeed), len(pairs))
	for _, p := range pairs {
		contractAddr := common.HexToAddress(p.Erc20Address)
		acc := network.App.AccountKeeper.GetAccount(ctx, contractAddr.Bytes())
		// no changes should be applied to native erc20s
		if p.ContractOwner != erc20types.OWNER_MODULE {
			require.Nil(t, acc)
			continue
		}
		ethAcct, ok := acc.(*types.EthAccount)
		require.True(t, ok)
		require.Equal(t, ethAcct.CodeHash, expCodeHash)
	}
}
