// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v18/utils"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

// erc20TokenPairHex is the string representation of the ERC-20 token pair address.
const erc20TokenPairHex = "0x80b5a32E4F032B2a058b4F29EC95EEfEEB87aDcd" //

// CreateGenesisWithTokenPairs creates a genesis that includes
// the WEVMOS and 'xmpl' token pairs
func CreateGenesisWithTokenPairs(keyring testkeyring.Keyring) network.CustomGenesisState {
	// Add all keys from the keyring to the genesis accounts as well.
	//
	// NOTE: This is necessary to enable the account to send EVM transactions,
	// because the Mono ante handler checks the account balance by querying the
	// account from the account keeper first. If these accounts are not in the genesis
	// state, the ante handler finds a zero balance because of the missing account.
	accs := keyring.GetAllAccAddrs()
	genesisAccounts := make([]*authtypes.BaseAccount, len(accs))
	for i, addr := range accs {
		genesisAccounts[i] = &authtypes.BaseAccount{
			Address:       addr.String(),
			PubKey:        nil,
			AccountNumber: uint64(i + 1),
			Sequence:      1,
		}
	}

	accGenesisState := authtypes.DefaultGenesisState()
	for _, genesisAccount := range genesisAccounts {
		// NOTE: This type requires to be packed into a *types.Any as seen on SDK tests,
		// e.g. https://github.com/evmos/cosmos-sdk/blob/v0.47.5-evmos.2/x/auth/keeper/keeper_test.go#L193-L223
		accGenesisState.Accounts = append(accGenesisState.Accounts, codectypes.UnsafePackAny(genesisAccount))
	}

	// Add token pairs to genesis
	erc20GenesisState := erc20types.DefaultGenesisState()
	erc20GenesisState.TokenPairs = []erc20types.TokenPair{{
		Erc20Address:  erc20TokenPairHex,
		Denom:         "xmpl",
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE, // NOTE: Owner is the module account since it's a native token and was registered through governance
	}, {
		Erc20Address:  erc20types.WEVMOSContractTestnet,
		Denom:         utils.BaseDenom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE, // NOTE: Owner is the module account since it's a native token and was registered through governance
	}}

	// STR v2: update the NativePrecompiles and DynamicPrecompiles
	// with the WEVMOS (default is mainnet) and 'xmpl' tokens in the erc20 params
	erc20GenesisState.Params.NativePrecompiles = []string{erc20types.WEVMOSContractTestnet}
	erc20GenesisState.Params.DynamicPrecompiles = []string{erc20TokenPairHex}

	// Add the smart contracts to the EVM genesis
	evmGenesisState := evmtypes.DefaultGenesisState()

	// Combine module genesis states
	return network.CustomGenesisState{
		authtypes.ModuleName:  accGenesisState,
		erc20types.ModuleName: erc20GenesisState,
		evmtypes.ModuleName:   evmGenesisState,
	}
}
