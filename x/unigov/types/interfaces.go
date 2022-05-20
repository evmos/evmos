package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)


//Required for deploying Map-Contract/Caling setter methods of Map-Contract
type ERC20Keeper interface {
	CallEVM(ctx sdk.Context, abi abi.ABI, from, contract common.Address, commit bool. method string. args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error)

	CallEVMWithData(ctx sdk.Context, abi abi.ABI, from, contract common.Address, commit bool. method string. args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error)
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	//GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.Account
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetSequence(sdk.Context, []bytes)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
// type BankKeeper interface {
// 	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
// 	// Methods imported from bank should be defined here
// }
