package types

import (
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cryptohd "github.com/evmos/evmos/v16/crypto/hd"
)

var _ keyring.Keyring = &itKeyring{}

type itKeyring struct {
	records []*itKeyringRecord
}

type itKeyringRecord struct {
	uid         string
	record      *keyring.Record
	testAccount *TestAccount
}

func NewIntegrationTestKeyring(accounts TestAccounts) keyring.Keyring {
	var records []*itKeyringRecord

	for i, account := range accounts {
		uid := fmt.Sprintf("it%d", i+1)

		anyPubKey, err := codectypes.NewAnyWithValue(account.GetPubKey())
		if err != nil {
			panic(err)
		}
		records = append(records, &itKeyringRecord{
			uid: uid,
			record: &keyring.Record{
				Name:   uid,
				PubKey: anyPubKey,
				Item:   nil,
			},
			testAccount: account,
		})
	}

	return &itKeyring{
		records: records,
	}
}

func (kr *itKeyring) Backend() string {
	return "test"
}

func (kr *itKeyring) List() ([]*keyring.Record, error) {
	var records []*keyring.Record
	for _, record := range kr.records {
		records = append(records, record.record)
	}

	return records, nil
}

func (kr *itKeyring) SupportedAlgorithms() (keyring.SigningAlgoList, keyring.SigningAlgoList) {
	return cryptohd.SupportedAlgorithms, cryptohd.SupportedAlgorithmsLedger
}

func (kr *itKeyring) Key(uid string) (record *keyring.Record, err error) {
	for _, keyringRecord := range kr.records {
		if keyringRecord.uid == uid {
			return keyringRecord.record, nil
		}
	}

	return nil, fmt.Errorf("keyring record with uid %s not found", uid)
}

func (kr *itKeyring) KeyByAddress(address sdk.Address) (*keyring.Record, error) {
	for _, record := range kr.records {
		if record.testAccount.GetCosmosAddress().Equals(address) {
			return record.record, nil
		}
	}

	return nil, fmt.Errorf("keyring record with address %s not found", address.String())
}

func (kr *itKeyring) Delete(uid string) error {
	var foundAndDeleted bool
	var newRecords []*itKeyringRecord
	for _, record := range kr.records {
		if record.uid == uid {
			foundAndDeleted = true
			continue
		}

		newRecords = append(newRecords, record)
	}
	if !foundAndDeleted {
		return fmt.Errorf("keyring record with uid %s not found", uid)
	}

	kr.records = newRecords
	return nil
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) DeleteByAddress(address sdk.Address) error {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) Rename(from string, to string) error {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) NewMnemonic(uid string, language keyring.Language, hdPath, bip39Passphrase string, algo keyring.SignatureAlgo) (*keyring.Record, string, error) {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) NewAccount(uid, mnemonic, bip39Passphrase, hdPath string, algo keyring.SignatureAlgo) (*keyring.Record, error) {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) SaveLedgerKey(uid string, algo keyring.SignatureAlgo, hrp string, coinType, account, index uint32) (*keyring.Record, error) {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) SaveOfflineKey(uid string, pubkey cryptotypes.PubKey) (*keyring.Record, error) {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) SaveMultisig(uid string, pubkey cryptotypes.PubKey) (*keyring.Record, error) {
	panic("implement me")
}

func (kr *itKeyring) Sign(uid string, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	for _, record := range kr.records {
		if record.uid == uid {
			return kr.sign(record, msg)
		}
	}

	return nil, nil, fmt.Errorf("keyring record with uid %s not found", uid)
}

func (kr *itKeyring) SignByAddress(address sdk.Address, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	for _, record := range kr.records {
		if record.testAccount.GetCosmosAddress().Equals(address) {
			return kr.sign(record, msg)
		}
	}

	return nil, nil, fmt.Errorf("keyring record with address %s not found", address.String())
}

func (kr *itKeyring) sign(record *itKeyringRecord, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	return record.testAccount.Signer.SignByAddress(record.testAccount.GetCosmosAddress(), msg)
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) ImportPrivKey(uid, armor, passphrase string) error {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) ImportPrivKeyHex(uid, privKey, algoStr string) error {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) ImportPubKey(uid string, armor string) error {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) ExportPubKeyArmor(uid string) (string, error) {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) ExportPubKeyArmorByAddress(address sdk.Address) (string, error) {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error) {
	panic("implement me")
}

//goland:noinspection GoUnusedParameter
func (kr *itKeyring) ExportPrivKeyArmorByAddress(address sdk.Address, encryptPassphrase string) (armor string, err error) {
	panic("implement me")
}

func (kr *itKeyring) MigrateAll() ([]*keyring.Record, error) {
	panic("implement me")
}
