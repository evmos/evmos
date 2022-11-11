package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// DefaultSendEnabled enabled
	DefaultSendEvmTxEnabled = true
	// DefaultReceiveEnabled enabled
	DefaultReceiveEvmTxEnabled = true
)

var (
	// KeySendEvmTxEnabled is store's key for SendEvmTxEnabled Params
	KeySendEvmTxEnabled = []byte("SendEvmTxEnabled")
	// KeyReceiveEnabled is store's key for ReceiveEvmTxEnabled Params
	KeyReceiveEvmTxEnabled = []byte("ReceiveEvmTxEnabled")
)

// NewParams creates a new parameter configuration for the ibc transfer module
func NewParams(enableSendEvmTxEnabled, enableReceiveEvmTxEnabled bool) Params {
	return Params{
		SendEvmTxEnabled:    enableSendEvmTxEnabled,
		ReceiveEvmTxEnabled: enableReceiveEvmTxEnabled,
	}
}

// DefaultParams is the default parameter configuration for the ibc-transfer module
func DefaultParams() Params {
	return NewParams(DefaultSendEvmTxEnabled, DefaultReceiveEvmTxEnabled)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeySendEvmTxEnabled, &p.SendEvmTxEnabled, validateBool),
		paramtypes.NewParamSetPair(KeyReceiveEvmTxEnabled, &p.ReceiveEvmTxEnabled, validateBool),
	}
}

func (p Params) Validate() error { return nil }
