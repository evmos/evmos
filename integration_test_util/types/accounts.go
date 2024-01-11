package types

//goland:noinspection SpellCheckingInspection
import (
	"crypto/ed25519"
	tmcrypto "github.com/cometbft/cometbft/crypto"
	tmed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v16/crypto/ethsecp256k1"
)

type TestAccount struct {
	PrivateKey *ethsecp256k1.PrivKey
	Signer     keyring.Signer
	Type       TestAccountType
}

type TestAccountType int8

const (
	TestAccountTypeValidator TestAccountType = iota
	TestAccountTypeWallet
)

func (a TestAccount) GetPubKey() cryptotypes.PubKey {
	return a.PrivateKey.PubKey()
}

func (a TestAccount) GetSdkPubKey() cryptotypes.PubKey {
	pk, err := cryptocodec.FromTmPubKeyInterface(a.GetTmPubKey())
	if err != nil {
		panic(err)
	}
	return pk
}

func (a TestAccount) GetTmPubKey() tmcrypto.PubKey {
	return a.GetTmPrivKey().PubKey()
}

func (a TestAccount) GetTmPrivKey() tmcrypto.PrivKey {
	//goland:noinspection GoDeprecation
	pv := mock.PV{
		PrivKey: &cosmosed25519.PrivKey{
			Key: ed25519.NewKeyFromSeed(a.PrivateKey.Key),
		},
	}

	var tmPrivEd25519 tmed25519.PrivKey
	tmPrivEd25519 = pv.PrivKey.Bytes()

	return tmPrivEd25519
}

// GetValidatorAddress returns validator address of the account, deliver from sdk pubkey.
// Should use suite.GetValidatorAddress() instead for correcting with Tendermint node mode.
func (a TestAccount) GetValidatorAddress() sdk.ValAddress {
	return sdk.ValAddress(a.GetPubKey().Address())
}

func (a TestAccount) GetConsensusAddress() sdk.ConsAddress {
	return sdk.ConsAddress(a.GetSdkPubKey().Address())
}

func (a TestAccount) GetCosmosAddress() sdk.AccAddress {
	return sdk.AccAddress(a.GetPubKey().Address())
}

func (a TestAccount) GetEthAddress() common.Address {
	return common.BytesToAddress(a.GetPubKey().Address().Bytes())
}

func (a TestAccount) ComputeContractAddress(nonce uint64) common.Address {
	return crypto.CreateAddress(a.GetEthAddress(), nonce)
}

type TestAccounts []*TestAccount

func (a TestAccounts) Number(num int) *TestAccount {
	return a[num-1]
}
