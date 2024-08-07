// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
<<<<<<< HEAD
	"github.com/ethereum/go-ethereum/common"
	testkeyring "github.com/evmos/evmos/v18/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v18/testutil/integration/evmos/network"
	utiltx "github.com/evmos/evmos/v18/testutil/tx"
	"github.com/evmos/evmos/v18/utils"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

// CreateGenesisWithTokenPairs creates a genesis that includes
// the WEVMOS and the provided denoms.
// If no denoms provided, creates only one dynamic precompile with the 'xmpl' denom.
func CreateGenesisWithTokenPairs(keyring testkeyring.Keyring, denoms ...string) network.CustomGenesisState {
=======
	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v19/utils"
	erc20types "github.com/evmos/evmos/v19/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

const (
	// erc20TokenPairHex is the string representation of the ERC-20 token pair address.
	erc20TokenPairHex = "0x80b5a32E4F032B2a058b4F29EC95EEfEEB87aDcd" //#nosec G101 -- these are not hardcoded credentials #gitleaks:allow
	// WEVMOSContractTestnet is the WEVMOS contract address for testnet
	WEVMOSContractTestnet = "0xcc491f589b45d4a3c679016195b3fb87d7848210"
)

func CreateGenesisWithTokenPairs(keyring testkeyring.Keyring) network.CustomGenesisState {
>>>>>>> main
	// Add all keys from the keyring to the genesis accounts as well.
	//
	// NOTE: This is necessary to enable the account to send EVM transactions,
	// because the Mono ante handler checks the account balance by querying the
	// account from the account keeper first. If these accounts are not in the genesis
	// state, the ante handler finds a zero balance because of the missing account.
<<<<<<< HEAD

	// if denom not provided, defaults to create only one dynamic erc20
	// precompile with the 'xmpl' denom
	if len(denoms) == 0 {
		denoms = []string{"xmpl"}
	}
=======
>>>>>>> main
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

<<<<<<< HEAD
	// get the common.Address to store the hex string addresses using EIP-55
	wevmosAddr := common.HexToAddress(erc20types.WEVMOSContractTestnet).Hex()

	// Add token pairs to genesis
	tokenPairs := make([]erc20types.TokenPair, 0, len(denoms)+1)
	// add token pair for fees token (wevmos)
	tokenPairs = append(tokenPairs, erc20types.TokenPair{
		Erc20Address:  wevmosAddr,
		Denom:         utils.BaseDenom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE, // NOTE: Owner is the module account since it's a native token and was registered through governance
	})

	dynPrecAddr := make([]string, 0, len(denoms))
	for _, denom := range denoms {
		addr := utiltx.GenerateAddress().Hex()
		tp := erc20types.TokenPair{
			Erc20Address:  addr,
			Denom:         denom,
			Enabled:       true,
			ContractOwner: erc20types.OWNER_MODULE, // NOTE: Owner is the module account since it's a native token and was registered through governance
		}
		tokenPairs = append(tokenPairs, tp)
		dynPrecAddr = append(dynPrecAddr, addr)
	}

	erc20GenesisState := erc20types.DefaultGenesisState()
	erc20GenesisState.TokenPairs = tokenPairs

	// STR v2: update the NativePrecompiles and DynamicPrecompiles
	// with the WEVMOS (default is mainnet) and 'xmpl' tokens in the erc20 params
	erc20GenesisState.Params.NativePrecompiles = []string{wevmosAddr}
	erc20GenesisState.Params.DynamicPrecompiles = dynPrecAddr
=======
	// Add token pairs to genesis
	erc20GenesisState := erc20types.DefaultGenesisState()
	erc20GenesisState.TokenPairs = []erc20types.TokenPair{{
		Erc20Address:  erc20TokenPairHex,
		Denom:         "xmpl",
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE, // NOTE: Owner is the module account since it's a native token and was registered through governance
	}, {
		Erc20Address:  WEVMOSContractTestnet,
		Denom:         utils.BaseDenom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE, // NOTE: Owner is the module account since it's a native token and was registered through governance
	}}
>>>>>>> main

	// Add the smart contracts to the EVM genesis
	evmGenesisState := evmtypes.DefaultGenesisState()

	// Combine module genesis states
	return network.CustomGenesisState{
		authtypes.ModuleName:  accGenesisState,
		erc20types.ModuleName: erc20GenesisState,
		evmtypes.ModuleName:   evmGenesisState,
	}
}
