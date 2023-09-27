package factory

import (
	sdkmath "cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

// CosmosTxArgs contains the params to create a cosmos tx
type CosmosTxArgs struct {
	// ChainID is the chain's id in cosmos format, e.g. 'evmos_9000-1'
	ChainID string
	// Gas to be used on the tx
	Gas uint64
	// GasPrice to use on tx
	GasPrice *sdkmath.Int
	// Fees is the fee to be used on the tx (amount and denom)
	Fees sdktypes.Coins
	// FeeGranter is the account address of the fee granter
	FeeGranter sdktypes.AccAddress
	// Msgs slice of messages to include on the tx
	Msgs []sdktypes.Msg
}

// CallArgs is a struct to define all relevant data to call a smart contract.
type CallArgs struct {
	// ContractABI is the ABI of the contract to call.
	ContractABI abi.ABI
	// MethodName is the name of the method to call.
	MethodName string
	// Args are the arguments to pass to the method.
	Args []interface{}
}
