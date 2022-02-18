package types

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewTestMsg generates a test message
func NewTestMsg(addrs ...sdk.AccAddress) *testdata.TestMsg {
	return testdata.NewTestMsg(addrs...)
}

// KeyTestPubAddr generates a test key pair
func KeyTestPubAddr() (cryptotypes.PrivKey, cryptotypes.PubKey, sdk.AccAddress) {
	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}
