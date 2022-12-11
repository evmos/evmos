package main_test

import (
	"crypto/ecdsa"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/api/cosmos/crypto/ed25519"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	auxTx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	ledgeraccounts "github.com/evmos/evmos-ledger-go/accounts"

	ethaccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/ethermint/ethereum/eip712"
	"github.com/evmos/ethermint/tests"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	ledger "github.com/evmos/evmos-ledger-go/ledger"

	ledgermocks "github.com/evmos/evmos-ledger-go/ledger/mocks"
	"github.com/evmos/evmos-ledger-go/usbwallet"

	"github.com/evmos/evmos/v10/app"
	. "github.com/onsi/ginkgo/v2"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	evm "github.com/evmos/ethermint/x/evm/types"
	claimstypes "github.com/evmos/evmos/v10/x/claims/types"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	. "github.com/onsi/gomega"
)

var s *LedgerTestSuite

type Ledger struct {
	hrp        string
	SECP256K1  ledger.EvmosSECP256K1
	mockWallet *ledgermocks.Wallet
	account    ledgeraccounts.Account

	privKey *ecdsa.PrivateKey
	pubKey  *ecdsa.PublicKey
}

type LedgerTestSuite struct {
	suite.Suite

	*Ledger

	ctx sdk.Context

	app            *app.Evmos
	validator      stakingtypes.Validator
	queryClient    bankTypes.QueryClient
	queryClientEvm evm.QueryClient

	ethAddr common.Address
	accAddr sdk.AccAddress
	signer  keyring.Signer

	consAddress sdk.ConsAddress
}

func TestLedger(t *testing.T) {
	s = &LedgerTestSuite{
		Ledger: &Ledger{},
	}
	suite.Run(t, s)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Evmosd Suite")
}

func (suite *LedgerTestSuite) SetupLedger() {
	suite.hrp = "evmos"

	hub, err := usbwallet.NewLedgerHub()
	suite.Require().NoError(err)

	mockWallet := new(ledgermocks.Wallet)
	suite.mockWallet = mockWallet
	suite.SECP256K1 = ledger.EvmosSECP256K1{Hub: hub, PrimaryWallet: mockWallet}

	suite.privKey, err = crypto.GenerateKey()
	suite.pubKey = &suite.privKey.PublicKey

	suite.Require().NoError(err)
	addr := crypto.PubkeyToAddress(*suite.pubKey)
	suite.account = ledgeraccounts.Account{
		Address:   addr,
		PublicKey: suite.pubKey,
	}
}

func (suite *LedgerTestSuite) RegisterMocks(signDocBytes []byte) {
	suite.mockWallet.On("Open", "").Return(nil)

	suite.mockWallet.On("Derive", ethaccounts.DefaultBaseDerivationPath, true).
		Return(suite.account, nil)

	suite.mockWallet.On("SignTypedData", suite.account, mock.AnythingOfType("TypedData")).Return().Run(func(args mock.Arguments) {
		fmt.Println("---------------------------")
		fmt.Printf("%t\n", args...)
		fmt.Println("---------------------------")
	})
}

func (s *LedgerTestSuite) SetupEvmosApp() {

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err, "can't generate private key")

	s.ethAddr = common.BytesToAddress(priv.PubKey().Address().Bytes())
	s.accAddr = sdk.AccAddress(s.ethAddr.Bytes())
	s.signer = tests.NewSigner(priv)

	// consensus kye
	privConsKey, err := ethsecp256k1.GenerateKey()
	s.Require().NoError(err, "can't generate private key")
	consAddress := sdk.ConsAddress(privConsKey.PubKey().Address())
	s.consAddress = consAddress

	eip712.SetEncodingConfig(encoding.MakeConfig(app.ModuleBasics))
	// init app
	s.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9001-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	// query clients
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	bankTypes.RegisterQueryServer(queryHelper, s.app.BankKeeper)
	s.queryClient = bankTypes.NewQueryClient(queryHelper)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, s.app.EvmKeeper)
	s.queryClientEvm = evm.NewQueryClient(queryHelperEvm)

	// bond denom
	params := claimstypes.DefaultParams()
	stakingParams := s.app.StakingKeeper.GetParams(s.ctx)
	stakingParams.BondDenom = params.GetClaimsDenom()
	s.app.StakingKeeper.SetParams(s.ctx, stakingParams)

	// Set Validator
	valAddr := sdk.ValAddress(s.ethAddr.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, privConsKey.PubKey(), stakingtypes.Description{})
	s.Require().NoError(err)

	validator = stakingkeeper.TestingUpdateValidator(s.app.StakingKeeper, s.ctx, validator, true)
	s.app.StakingKeeper.AfterValidatorCreated(s.ctx, validator.GetOperator())
	err = s.app.StakingKeeper.SetValidatorByConsAddr(s.ctx, validator)
	s.Require().NoError(err)

	s.validator = validator
}

func (suite *LedgerTestSuite) getMockTxAmino() []byte {
	tmp := `{"account_number":"0","chain_id":"evmos_9000-1","fee":{"amount":[{"amount":"150","denom":"aevmos"}],"gas":"20000"},"memo":"memo","msgs":[{"type":"cosmos-sdk/MsgSend","value":{"amount":[{"amount":"150","denom":"aevmos"}],"from_address":"evmos1r5sckdd808qvg7p8d0auaw896zcluqfd7djffp","to_address":"evmos10t8ca2w09ykd6ph0agdz5stvgau47whhaggl9a"}}],"sequence":"6"}`
	return []byte(tmp)
}

func (suite *LedgerTestSuite) getMockTxProtobuf(toAddr sdk.AccAddress, amount int64) []byte {
	marshaler := codec.NewProtoCodec(codecTypes.NewInterfaceRegistry())

	memo := "memo"
	msg := bankTypes.NewMsgSend(
		s.accAddr,
		toAddr,
		[]types.Coin{
			{
				Denom:  "aevmos",
				Amount: types.NewIntFromUint64(1000),
			},
		},
	)

	msgAsAny, err := codecTypes.NewAnyWithValue(msg)
	suite.Require().NoError(err)

	body := &txTypes.TxBody{
		Messages: [](*codecTypes.Any){
			msgAsAny,
		},
		Memo: memo,
	}

	pkBytes := crypto.FromECDSAPub(suite.pubKey)
	edPubKey := &ed25519.PubKey{Key: pkBytes}
	pubKeyAsAny, err := codecTypes.NewAnyWithValue(edPubKey)
	suite.Require().NoError(err)

	signingMode := txTypes.ModeInfo_Single_{
		Single: &txTypes.ModeInfo_Single{
			Mode: signing.SignMode_SIGN_MODE_DIRECT,
		},
	}

	signerInfo := &txTypes.SignerInfo{
		PublicKey: pubKeyAsAny,
		ModeInfo: &txTypes.ModeInfo{
			Sum: &signingMode,
		},
		Sequence: 6,
	}

	fee := txTypes.Fee{Amount: types.NewCoins(types.NewInt64Coin("aevmos", amount)), GasLimit: 20000}

	authInfo := &txTypes.AuthInfo{
		SignerInfos: [](*txTypes.SignerInfo){signerInfo},
		Fee:         &fee,
	}

	bodyBytes := marshaler.MustMarshal(body)
	authInfoBytes := marshaler.MustMarshal(authInfo)

	signBytes, err := auxTx.DirectSignBytes(
		bodyBytes,
		authInfoBytes,
		"evmos_9000-1",
		0,
	)
	suite.Require().NoError(err)

	return signBytes
}
