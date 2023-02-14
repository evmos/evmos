package keeper_test

import (
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	. "github.com/onsi/gomega"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/contracts"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/encoding"
	"github.com/evmos/evmos/v11/server/config"
	"github.com/evmos/evmos/v11/tests"
	"github.com/evmos/evmos/v11/testutil"
	evmostypes "github.com/evmos/evmos/v11/types"
	"github.com/evmos/evmos/v11/utils"
	"github.com/evmos/evmos/v11/x/claims/types"
	evm "github.com/evmos/evmos/v11/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v11/x/feemarket/types"
	incentivestypes "github.com/evmos/evmos/v11/x/incentives/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
)

func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = tests.NewSigner(priv)

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
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

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.ClaimsKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelperEvm)

	params := types.DefaultParams()
	params.AirdropStartTime = suite.ctx.BlockTime().UTC()
	err = suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
	require.NoError(t, err)

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = params.GetClaimsDenom()
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, privCons.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	validators := s.app.StakingKeeper.GetValidators(s.ctx, 1)
	suite.validator = validators[0]

	suite.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())
}

func (suite *KeeperTestSuite) SetupTestWithEscrow() {
	suite.SetupTest()
	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)

	coins := sdk.NewCoins(sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(10000000)))
	err := testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, types.ModuleName, coins)
	suite.Require().NoError(err)
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	header := suite.ctx.BlockHeader()
	suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{Height: header.Height})
	_ = suite.app.Commit()

	header.Height++
	header.Time = header.Time.Add(t)
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.ClaimsKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelperEvm)
}

func newEthAccount(baseAccount *authtypes.BaseAccount) evmostypes.EthAccount {
	return evmostypes.EthAccount{
		BaseAccount: baseAccount,
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}
}

func getAddr(priv *ethsecp256k1.PrivKey) sdk.AccAddress {
	return sdk.AccAddress(priv.PubKey().Address().Bytes())
}

func delegate(priv *ethsecp256k1.PrivKey, delegateAmount sdk.Coin) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	val, err := sdk.ValAddressFromBech32(s.validator.OperatorAddress)
	s.Require().NoError(err)

	delegateMsg := stakingtypes.NewMsgDelegate(accountAddress, val, delegateAmount)
	deliverTx(priv, delegateMsg)
}

func govProposal(priv *ethsecp256k1.PrivKey) uint64 {
	stakeDenom := stakingtypes.DefaultParams().BondDenom
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	contractAddress := deployContract(priv)
	content := incentivestypes.NewRegisterIncentiveProposal(
		"test",
		"description",
		contractAddress.String(),
		sdk.DecCoins{sdk.NewDecCoinFromDec(utils.BaseDenom, sdk.NewDecWithPrec(5, 2))},
		1000,
	)

	deposit := sdk.NewCoins(sdk.NewCoin(stakeDenom, sdk.NewInt(100000000)))
	msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, accountAddress)
	s.Require().NoError(err)

	res := deliverTx(priv, msg)
	submitEvent := res.GetEvents()[8]
	Expect(submitEvent.Type).To(Equal("submit_proposal"))
	Expect(string(submitEvent.Attributes[0].Key)).To(Equal("proposal_id"))

	proposalID, err := strconv.ParseUint(string(submitEvent.Attributes[0].Value), 10, 64)
	s.Require().NoError(err)

	return proposalID
}

func vote(priv *ethsecp256k1.PrivKey, proposalID uint64) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	voteMsg := govv1beta1.NewMsgVote(accountAddress, proposalID, govv1beta1.OptionAbstain)
	deliverTx(priv, voteMsg)
}

func sendEthToSelf(priv *ethsecp256k1.PrivKey) {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	msgEthereumTx := evm.NewTx(chainID, nonce, &from, nil, 100000, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), nil, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()
	performEthTx(priv, msgEthereumTx)
}

func deployContract(priv *ethsecp256k1.PrivKey) common.Address {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "Test", "TTT", uint8(18))
	s.Require().NoError(err)

	data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &s.address,
		Data: (*hexutil.Bytes)(&data),
	})
	s.Require().NoError(err)

	ctx := sdk.WrapSDKContext(s.ctx)
	res, err := s.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: config.DefaultGasCap,
	})
	s.Require().NoError(err)

	msgEthereumTx := evm.NewTxContract(chainID, nonce, nil, res.Gas, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), data, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()

	performEthTx(priv, msgEthereumTx)
	s.Commit()

	contractAddress := crypto.CreateAddress(from, nonce)
	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)
	s.Require().NotEmpty(acc)
	s.Require().True(acc.IsContract())
	return contractAddress
}

func performEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evm.MsgEthereumTx) {
	// Sign transaction
	err := msgEthereumTx.Sign(s.ethSigner, tests.NewSigner(priv))
	s.Require().NoError(err)

	// Assemble transaction from fields
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	tx, err := msgEthereumTx.BuildTx(txBuilder, utils.BaseDenom)
	s.Require().NoError(err)

	// Encode transaction by default Tx encoder and broadcasted over the network
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(tx)
	s.Require().NoError(err)

	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	Expect(res.IsOK()).To(Equal(true))
}

func deliverTx(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(1000000)
	txBuilder.SetFeeAmount(sdk.Coins{defaultTxFee})

	err := txBuilder.SetMsgs(msgs...)
	s.Require().NoError(err)

	seq, err := s.app.AccountKeeper.GetSequence(s.ctx, accountAddress)
	s.Require().NoError(err)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := s.app.AccountKeeper.GetAccount(s.ctx, accountAddress).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       s.ctx.ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		seq,
	)
	s.Require().NoError(err)

	sigsV2 = []signing.SignatureV2{sigV2}
	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	Expect(res.IsOK()).To(Equal(true), res.Log)
	return res
}

func getEthTxFee() sdk.Coin {
	baseFee := s.app.FeeMarketKeeper.GetBaseFee(s.ctx)
	baseFee.Mul(baseFee, big.NewInt(100000))
	feeAmt := baseFee.Quo(baseFee, big.NewInt(2))
	return sdk.NewCoin(utils.BaseDenom, sdkmath.NewIntFromBigInt(feeAmt))
}
