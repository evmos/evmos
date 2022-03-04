package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

// Parameter store key
var (
	ParamStoreKeyEnableWithdraw  = []byte("EnableWithdraw")
	ParamStoreKeyEnabledChannels = []byte("EnabledChannels")
)

// DefaultChannels defines the list of default IBC channels that can withdraw
// stuck funds to users
var DefaultChannels = []string{
	"channel-0", // osmosis
	"channel-1", // umee
	"channel-3", // cosmos hub
}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	enableWithdraw bool,
	enabledChannels ...string,
) Params {
	return Params{
		EnableWithdraw:  enableWithdraw,
		EnabledChannels: enabledChannels,
	}
}

// DefaultParams defines the default params for the withdraw module
func DefaultParams() Params {
	return Params{
		EnableWithdraw:  true,
		EnabledChannels: DefaultChannels,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableWithdraw, &p.EnableWithdraw, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyEnabledChannels, &p.EnabledChannels, validateChannels),
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateChannels(i interface{}) error {
	channels, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, channel := range channels {
		if !channeltypes.IsValidChannelID(channel) {
			return fmt.Errorf("invalid channel id %s", channel)
		}
	}

	return nil
}

// IsChannelEnabled returns true if the channel provided is in the list of
// enabled channels
func (p Params) IsChannelEnabled(channel string) bool {
	for _, enabledChannel := range p.EnabledChannels {
		if channel == enabledChannel {
			return true
		}
	}

	return false
}

// Validate checks that the fields have valid values
func (p Params) Validate() error {
	return validateChannels(p.EnabledChannels)
}
