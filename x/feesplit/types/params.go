package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

// Parameter store key
var (
	DefaultEnableFeeSplit  = true
	DefaultDeveloperShares = sdk.NewDecWithPrec(50, 2) // 50%
	DefaultFeeDiscount     = sdk.NewDecWithPrec(50, 2) // 50%
	// Cost for executing `crypto.CreateAddress` must be at least 36 gas for the
	// contained keccak256(word) operation
	DefaultAddrDerivationCostCreate = uint64(50)

	ParamStoreKeyEnableFeeSplit           = []byte("EnableFeeSplit")
	ParamStoreKeyDeveloperShares          = []byte("DeveloperShares")
	ParamStoreKeyAddrDerivationCostCreate = []byte("AddrDerivationCostCreate")
	// ParamsStoreKeyFeeDiscount is the store key for the FeeDiscount parameter
	ParamsStoreKeyFeeDiscount = []byte("FeeDiscount")
	// ParamsStoreKeyEligibleMessages is the store key for the EligibleMessages parameter
	ParamsStoreKeyEligibleMessages = []byte("EligibleMessages")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(
	enableFeeSplit bool,
	developerShares,
	feeDiscount sdk.Dec,
	addrDerivationCostCreate uint64,
	eligibleMessages ...string,
) Params {
	return Params{
		EnableFeeSplit:           enableFeeSplit,
		DeveloperShares:          developerShares,
		AddrDerivationCostCreate: addrDerivationCostCreate,
		FeeDiscount:              feeDiscount,
		EligibleMessages:         eligibleMessages,
	}
}

func DefaultParams() Params {
	return Params{
		EnableFeeSplit:           DefaultEnableFeeSplit,
		DeveloperShares:          DefaultDeveloperShares,
		AddrDerivationCostCreate: DefaultAddrDerivationCostCreate,
		FeeDiscount:              DefaultFeeDiscount,
		EligibleMessages: []string{
			sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
			sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
			sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
			sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeout{}),
		},
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableFeeSplit, &p.EnableFeeSplit, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyDeveloperShares, &p.DeveloperShares, validateShares),
		paramtypes.NewParamSetPair(ParamStoreKeyAddrDerivationCostCreate, &p.AddrDerivationCostCreate, validateUint64),
		paramtypes.NewParamSetPair(ParamsStoreKeyFeeDiscount, &p.FeeDiscount, validateShares),
		paramtypes.NewParamSetPair(ParamsStoreKeyEligibleMessages, &p.EligibleMessages, validateTypeURLs),
	}
}

// IsEligibleMsg iterates over the messages and returns true if the message is
// eligible for a discount.
func (p Params) IsEligibleMsg(msgURL string) bool {
	for _, t := range p.EligibleMessages {
		if t == msgURL {
			return true
		}
	}
	return false
}

func validateUint64(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateShares(i interface{}) error {
	v, ok := i.(sdk.Dec)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("invalid parameter: nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("value cannot be negative: %T", i)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("value cannot be greater than 1: %T", i)
	}

	return nil
}

func validateTypeURLs(i interface{}) error {
	v, ok := i.([]string)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, url := range v {
		if strings.TrimSpace(url) == "" {
			return fmt.Errorf("invalid type URL: %s", url)
		}
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateBool(p.EnableFeeSplit); err != nil {
		return err
	}
	if err := validateShares(p.DeveloperShares); err != nil {
		return err
	}
	if err := validateShares(p.FeeDiscount); err != nil {
		return err
	}
	if err := validateTypeURLs(p.EligibleMessages); err != nil {
		return err
	}
	return validateUint64(p.AddrDerivationCostCreate)
}
