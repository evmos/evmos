package blocksdk_test

import (
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v14/app"
	blocksdk "github.com/evmos/evmos/v14/app/block-sdk"
	"github.com/evmos/evmos/v14/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v14/encoding"
	"github.com/evmos/evmos/v14/testutil"
	testutiltx "github.com/evmos/evmos/v14/testutil/tx"
	evmostypes "github.com/evmos/evmos/v14/types"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
	"github.com/stretchr/testify/suite"
)

type SignerExtractionAdapterTestSuite struct {
	suite.Suite

	txConfig client.TxConfig
	pk       cryptotypes.PrivKey
	app      *app.Evmos
	ctx      sdk.Context

	extractor blocksdk.SignerExtractionAdapter
	signer    sdk.AccAddress
	account   *evmostypes.EthAccount
}

const (
	privKey = "b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291"
	chainID = "evmos_6102-1"
)

func TestSignerExtractionAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(SignerExtractionAdapterTestSuite))
}

func (s *SignerExtractionAdapterTestSuite) SetupTest() {
	// create the encoding config
	s.txConfig = encoding.MakeConfig(app.ModuleBasics).TxConfig
	// create the app
	s.app = app.Setup(true, nil, chainID)

	// create a private key (eth)
	ecdsaPriv, err := crypto.HexToECDSA(privKey)
	s.Require().NoError(err)
	s.pk = &ethsecp256k1.PrivKey{
		Key: crypto.FromECDSA(ecdsaPriv),
	}

	address := common.BytesToAddress(s.pk.PubKey().Address().Bytes())

	s.signer = sdk.AccAddress(address.Bytes())

	// set the account to state
	s.account = &evmostypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(s.signer, nil, 0, 0),
	}

	s.ctx = s.app.NewContext(true, testutil.NewHeader(1, time.Now().UTC(), chainID, sdk.ConsAddress([]byte{}), []byte{}, []byte{}))

	s.app.AccountKeeper.SetAccount(s.ctx, s.account)

	s.extractor = blocksdk.NewSignerExtractorAdapter()
}

// test that the signers of a normal cosmos-tx can be extracted
func (s *SignerExtractionAdapterTestSuite) TestCosmosTx() {
	cosmosTxArgs := testutiltx.CosmosTxArgs{
		TxCfg:    s.txConfig,
		Priv:     s.pk,
		ChainID:  chainID,
		Gas:      100000,
		GasPrice: nil,
		Fees:     sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(100))),
		Msgs: []sdk.Msg{
			&banktypes.MsgSend{
				FromAddress: "a",
				ToAddress:   "b",
				Amount:      sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(100))),
			},
		},
	}

	s.Run("extract signers from a regularly signed cosmos-tx", func() {
		// update the sequence
		s.app.AccountKeeper.SetAccount(
			s.ctx,
			&evmostypes.EthAccount{
				BaseAccount: authtypes.NewBaseAccount(s.signer, nil, 0, 1),
			},
		)

		// create a cosmos tx signed by the above signer
		tx, err := testutiltx.PrepareCosmosTx(
			s.ctx,
			s.app,
			cosmosTxArgs,
		)
		s.Require().NoError(err)

		signers, err := s.extractor.GetSigners(tx)
		s.Require().NoError(err)

		s.Require().Len(signers, 1)
		s.Require().Equal(signers[0].Signer, s.signer)
		s.Require().Equal(signers[0].Sequence, uint64(1))
	})

	s.Run("extract the signers from an EIP-712 signed cosmos-tx", func() {
		// update the sequence again
		s.app.AccountKeeper.SetAccount(
			s.ctx,
			&evmostypes.EthAccount{
				BaseAccount: authtypes.NewBaseAccount(s.signer, nil, 0, 2),
			},
		)

		// create a cosmos tx signed by the above signer
		tx, err := testutiltx.CreateEIP712CosmosTx(
			s.ctx,
			s.app,
			testutiltx.EIP712TxArgs{
				CosmosTxArgs: cosmosTxArgs,
			},
		)
		s.Require().NoError(err)

		signers, err := s.extractor.GetSigners(tx)
		s.Require().NoError(err)

		s.Require().Len(signers, 1)
		s.Require().Equal(signers[0].Signer, s.signer)
		s.Require().Equal(signers[0].Sequence, uint64(2))
	})
}

// test that the signers of an ethereum tx can be extracted
func (s *SignerExtractionAdapterTestSuite) TestEthereumSignature() {
	// increment sequence
	s.app.AccountKeeper.SetAccount(
		s.ctx,
		&evmostypes.EthAccount{
			BaseAccount: authtypes.NewBaseAccount(s.signer, nil, 0, 3),
		},
	)

	msgEthTx := evmtypes.NewTx(&evmtypes.EvmTxArgs{
		ChainID: big.NewInt(6102),
		Nonce:   3,
		To:      nil,
	})

	msgEthTx.From = s.account.EthAddress().String()

	tx, err := testutiltx.PrepareEthTx(
		s.txConfig,
		s.app,
		s.pk,
		msgEthTx,
	)
	s.Require().NoError(err)

	signers, err := s.extractor.GetSigners(tx)
	s.Require().NoError(err)

	s.Require().Len(signers, 1)
	s.Require().Equal(signers[0].Signer, s.signer)
	s.Require().Equal(signers[0].Sequence, uint64(3))
}
