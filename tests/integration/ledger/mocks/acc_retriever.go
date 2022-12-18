package mocks

import (
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Account interface {
	GetAddress() sdk.AccAddress
	GetPubKey() cryptotypes.PubKey // can return nil.
	GetAccountNumber() uint64
	GetSequence() uint64
}

// MockAccountRetriever defines a no-op basic AccountRetriever that can be used
// in mocked contexts. Tests or context that need more sophisticated testing
// state should implement their own mock AccountRetriever.
type MockAccountRetriever struct{}

func (mar MockAccountRetriever) GetAccount(_ client.Context, _ sdk.AccAddress) (Account, error) {
	return nil, nil
}

func (mar MockAccountRetriever) GetAccountWithHeight(_ client.Context, _ sdk.AccAddress) (Account, int64, error) {
	return nil, 0, nil
}

func (mar MockAccountRetriever) EnsureExists(_ client.Context, _ sdk.AccAddress) error {
	return nil
}

func (mar MockAccountRetriever) GetAccountNumberSequence(_ client.Context, _ sdk.AccAddress) (uint64, uint64, error) {
	return 0, 0, nil
}
