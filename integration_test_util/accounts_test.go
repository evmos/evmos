package integration_test_util

//goland:noinspection SpellCheckingInspection
import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

//goland:noinspection SpellCheckingInspection
func Test_TestAccount_KeysAndAddresses(t *testing.T) {
	t.Run("test validator account", func(t *testing.T) {
		valAcc := newTestAccountFromMnemonic(t, IT_VAL_1_MNEMONIC)

		require.Equal(t, IT_VAL_1_VAL_ADDR, valAcc.GetValidatorAddress().String())

		require.Equal(t, IT_VAL_1_CONS_ADDR, valAcc.GetConsensusAddress().String())

		require.Equal(t, IT_VAL_1_ADDR, valAcc.GetCosmosAddress().String())

		require.Equal(
			t,
			"EthPubKeySecp256k1{0213B178097B00B5E87CD81D4F02E0F6DBF4B10608CE4B409A630C7E837F973350}",
			valAcc.GetPubKey().String(),
		)

		require.Equal(
			t,
			"PubKeyEd25519{EC78352B42A13C1938886E4242AD7F2DD39DA55B886ECA06B826D50008C7C086}",
			valAcc.GetSdkPubKey().String(),
		)

		const tmPub = "6323691963BFA43CB659BDA8676B43D2B3CCCE10"
		require.Equal(
			t,
			tmPub,
			valAcc.GetTmPubKey().Address().String(),
		)
		tmPrivKey := valAcc.GetTmPrivKey()
		require.Equal(
			t,
			tmPub,
			tmPrivKey.PubKey().Address().String(),
		)

		require.Equal(
			t,
			"671959de1da2c3860f59ee81c19eeba23f18f585e22c47e3dfe559970d473038ec78352b42a13c1938886e4242ad7f2dd39da55b886eca06b826d50008c7c086",
			hex.EncodeToString(tmPrivKey.Bytes()),
		)
	})

	t.Run("test wallet account", func(t *testing.T) {
		walAcc := newTestAccountFromMnemonic(t, IT_WAL_1_MNEMONIC)

		require.Equal(t, IT_WAL_1_ETH_ADDR, walAcc.GetEthAddress().String())

		require.Equal(t, IT_WAL_1_ADDR, walAcc.GetCosmosAddress().String())

		require.Equal(
			t,
			"EthPubKeySecp256k1{020826F89EFBC5E5CFC8930E9E976D12EFA79821BF89AF73B9DAE6DE41ACEF6DB3}",
			walAcc.GetPubKey().String(),
		)

		require.Equal(
			t,
			"PubKeyEd25519{7570E4698A126B47D5E2698C74ED728F2DE06593AE72E1987FA25BCFCD46A4BA}",
			walAcc.GetSdkPubKey().String(),
		)
	})
}
