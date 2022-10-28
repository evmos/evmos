package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// ModuleCdc references the global incentives module codec. Note, the codec
// should ONLY be used in certain instances of tests and for JSON encoding.
//
// The actual codec used for serialization should be provided to
// modules/incentives and defined at the application level.
var ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*govv1beta1.Content)(nil),
		&RegisterIncentiveProposal{},
		&CancelIncentiveProposal{},
	)
}
