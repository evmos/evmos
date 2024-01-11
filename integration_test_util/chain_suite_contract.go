package integration_test_util

//goland:noinspection SpellCheckingInspection
import (
	sdkmath "cosmossdk.io/math"
	_ "embed" // embed compiled smart contract
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v16/contracts"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"math"
	"math/big"
)

var (
	//go:embed compiled_contracts/1-storage.json
	Contract1StorageJson []byte

	// Contract1Storage is the compiled storage contract
	Contract1Storage evmtypes.CompiledContract

	//go:embed compiled_contracts/2-wevmos.json
	Contract2WEvmosJson []byte

	// Contract2WEvmos is the compiled WEVMOS contract
	Contract2WEvmos evmtypes.CompiledContract

	//go:embed compiled_contracts/3-nft721.json
	Contract3Nft721Json []byte

	// Contract3Nft721 is the compiled NFT-721 contract
	Contract3Nft721 evmtypes.CompiledContract

	//go:embed compiled_contracts/4-nft1155.json
	Contract4Nft1155Json []byte

	// Contract4Nft1155 is the compiled NFT-1155 contract
	Contract4Nft1155 evmtypes.CompiledContract

	//go:embed compiled_contracts/5-create-Foo.json
	Contract5CreateFooJson []byte
	//go:embed compiled_contracts/5-create-Bar.json
	Contract5CreateBarJson []byte
	//go:embed compiled_contracts/5-create-BarInteraction.json
	Contract5CreateBarInteractionJson []byte

	// Contract5CreateFooContract is the compiled Foo-contract on 5-create.sol
	Contract5CreateFooContract evmtypes.CompiledContract
	// Contract5CreateBarContract is the compiled Bar-contract on 5-create.sol
	Contract5CreateBarContract evmtypes.CompiledContract
	// Contract5CreateBarInteractionContract is the compiled BarInteraction-contract on 5-create.sol
	Contract5CreateBarInteractionContract evmtypes.CompiledContract
)

func init() {
	var err error

	// initialize embedded compiled contracts

	// 1-storage.sol

	err = json.Unmarshal(Contract1StorageJson, &Contract1Storage)
	if err != nil {
		panic(err)
	}

	if len(Contract1Storage.Bin) == 0 {
		panic("load contract failed")
	}

	// 2-wevmos.sol

	err = json.Unmarshal(Contract2WEvmosJson, &Contract2WEvmos)
	if err != nil {
		panic(err)
	}

	if len(Contract2WEvmos.Bin) == 0 {
		panic("load contract failed")
	}

	// 3-nft721.sol

	err = json.Unmarshal(Contract3Nft721Json, &Contract3Nft721)
	if err != nil {
		panic(err)
	}

	if len(Contract3Nft721.Bin) == 0 {
		panic("load contract failed")
	}

	// 4-nft1155.sol

	err = json.Unmarshal(Contract4Nft1155Json, &Contract4Nft1155)
	if err != nil {
		panic(err)
	}

	if len(Contract4Nft1155.Bin) == 0 {
		panic("load contract failed")
	}

	// 5-create.sol

	err = json.Unmarshal(Contract5CreateFooJson, &Contract5CreateFooContract)
	if err != nil {
		panic(err)
	}

	if len(Contract5CreateFooContract.Bin) == 0 {
		panic("load contract failed")
	}

	err = json.Unmarshal(Contract5CreateBarJson, &Contract5CreateBarContract)
	if err != nil {
		panic(err)
	}

	if len(Contract5CreateBarContract.Bin) == 0 {
		panic("load contract failed")
	}

	err = json.Unmarshal(Contract5CreateBarInteractionJson, &Contract5CreateBarInteractionContract)
	if err != nil {
		panic(err)
	}

	if len(Contract5CreateBarInteractionContract.Bin) == 0 {
		panic("load contract failed")
	}
}

// TxDeployErc20Contract deploys a new ERC20 contract with the given name, symbol and decimals.
// The given deployer will be used to deploy the contract.
func (suite *ChainIntegrationTestSuite) TxDeployErc20Contract(deployer *itutiltypes.TestAccount, name, symbol string, decimals uint8) (common.Address, *evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(deployer)

	return suite.TxDeployContract(
		suite.CurrentContext,
		deployer,
		contracts.ERC20MinterBurnerDecimalsContract,
		name, symbol, decimals,
	)
}

// TxDeploy1StorageContract deploys the embedded pre-compiled contract "1-storage.sol".
func (suite *ChainIntegrationTestSuite) TxDeploy1StorageContract(deployer *itutiltypes.TestAccount) (common.Address, *evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(deployer)

	return suite.TxDeployContract(
		suite.CurrentContext,
		deployer,
		Contract1Storage,
	)
}

// TxDeploy2WEvmosContract deploys the embedded pre-compiled contract "2-wevmos.sol".
func (suite *ChainIntegrationTestSuite) TxDeploy2WEvmosContract(deployer, rich *itutiltypes.TestAccount) (common.Address, *evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(deployer)

	if rich == nil {
		rich = deployer
	}

	return suite.TxDeployContract(
		suite.CurrentContext,
		deployer,
		Contract2WEvmos,
		rich.GetEthAddress(),
	)
}

// TxDeploy3Nft721Contract deploys the embedded pre-compiled contract "3-nft721.sol".
func (suite *ChainIntegrationTestSuite) TxDeploy3Nft721Contract(deployer, rich *itutiltypes.TestAccount) (common.Address, *evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(deployer)

	if rich == nil {
		rich = deployer
	}

	return suite.TxDeployContract(
		suite.CurrentContext,
		deployer,
		Contract3Nft721,
		rich.GetEthAddress(),
	)
}

// TxDeploy4Nft1155Contract deploys the embedded pre-compiled contract "4-nft1155.sol".
func (suite *ChainIntegrationTestSuite) TxDeploy4Nft1155Contract(deployer, rich *itutiltypes.TestAccount) (common.Address, *evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(deployer)

	if rich == nil {
		rich = deployer
	}

	return suite.TxDeployContract(
		suite.CurrentContext,
		deployer,
		Contract4Nft1155,
		rich.GetEthAddress(),
	)
}

// TxDeploy5CreateFooContract deploys the Foo contract within the embedded pre-compiled contract "5-create.sol".
func (suite *ChainIntegrationTestSuite) TxDeploy5CreateFooContract(deployer *itutiltypes.TestAccount) (common.Address, *evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(deployer)

	return suite.TxDeployContract(
		suite.CurrentContext,
		deployer,
		Contract5CreateFooContract,
	)
}

// TxDeploy5CreateBarContract deploys the Bar contract within the embedded pre-compiled contract "5-create.sol".
func (suite *ChainIntegrationTestSuite) TxDeploy5CreateBarContract(deployer *itutiltypes.TestAccount) (common.Address, *evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(deployer)

	return suite.TxDeployContract(
		suite.CurrentContext,
		deployer,
		Contract5CreateBarContract,
	)
}

// TxDeploy5CreateBarInteractionContract deploys the BarInteraction contract within the embedded pre-compiled contract "5-create.sol".
func (suite *ChainIntegrationTestSuite) TxDeploy5CreateBarInteractionContract(deployer *itutiltypes.TestAccount, contractBarAddress common.Address) (common.Address, *evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(deployer)

	addr, evmTx, res, err := suite.TxDeployContract(
		suite.CurrentContext,
		deployer,
		Contract5CreateBarInteractionContract,
	)
	suite.Commit()
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	data, err := Contract5CreateBarInteractionContract.ABI.Pack("setBarAddr", contractBarAddress)
	suite.Require().NoError(err)
	_, _, err = suite.TxSendEvmTx(suite.CurrentContext, deployer, &addr, nil, data)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	return addr, evmTx, res, nil
}

// TxDeployContract deploys the given compiled-contract with the given constructor arguments, using the given deployer.
func (suite *ChainIntegrationTestSuite) TxDeployContract(ctx sdk.Context, deployer *itutiltypes.TestAccount, contract evmtypes.CompiledContract, constructorArgs ...interface{}) (common.Address, *evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(deployer)

	nonce := suite.ChainApp.EvmKeeper().GetNonce(ctx, deployer.GetEthAddress())

	ctorArgs, err := contract.ABI.Pack("", constructorArgs...)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	data := append(contract.Bin, ctorArgs...)

	evmTx, resDeliver, err := suite.TxSendEvmTx(ctx, deployer, nil, nil, data)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	return crypto.CreateAddress(deployer.GetEthAddress(), nonce), evmTx, resDeliver, nil
}

// TxMintErc20Token call the "mint" function of the given ERC-20 contract, minting token to given account.
// Minter account will be used to call the function.
func (suite *ChainIntegrationTestSuite) TxMintErc20Token(contract common.Address, minter, mintTo *itutiltypes.TestAccount, amount uint16, decimals uint8) (*evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(minter)
	suite.Require().NotNil(mintTo)

	mintAmt := computeAmount(amount, decimals)
	data, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("mint", mintTo.GetEthAddress(), mintAmt.BigInt())
	suite.Require().NoError(err)
	return suite.TxSendEvmTx(suite.CurrentContext, minter, &contract, nil, data)
}

// TxTransferErc20Token call the "transfer" function of the given ERC20 contract, transferring token from sender to receiver.
// Sender account will be used to call the function.
func (suite *ChainIntegrationTestSuite) TxTransferErc20Token(contract common.Address, sender, receiver *itutiltypes.TestAccount, amount uint16, decimals uint8) (*evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(sender)
	suite.Require().NotNil(receiver)

	data := suite.prepareTransferErc20TokenData(receiver, amount, decimals)
	return suite.TxSendEvmTx(suite.CurrentContext, sender, &contract, nil, data)
}

// TxTransferErc20TokenAsync is the same as TxTransferErc20Token but with Async delivery mode.
func (suite *ChainIntegrationTestSuite) TxTransferErc20TokenAsync(contract common.Address, sender, receiver *itutiltypes.TestAccount, amount uint16, decimals uint8) (*evmtypes.MsgEthereumTx, error) {
	suite.Require().NotNil(sender)
	suite.Require().NotNil(receiver)

	data := suite.prepareTransferErc20TokenData(receiver, amount, decimals)
	return suite.TxSendEvmTxAsync(suite.CurrentContext, sender, &contract, nil, data)
}

// prepareTransferErc20TokenData computes the call data for the "transfer" function of the given ERC-20 contract, with given amount of token to transfer.
func (suite *ChainIntegrationTestSuite) prepareTransferErc20TokenData(receiver *itutiltypes.TestAccount, amount uint16, decimals uint8) []byte {
	suite.Require().NotNil(receiver)
	transferAmt := computeAmount(amount, decimals)
	data, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("transfer", receiver.GetEthAddress(), transferAmt.BigInt())
	suite.Require().NoError(err)
	return data
}

// TxTransferNft721Token call the "safeTransferFrom" function of the given ERC-721 contract, transferring a NFT token from sender to receiver.
// Sender account will be used to call the function.
func (suite *ChainIntegrationTestSuite) TxTransferNft721Token(contract common.Address, abi abi.ABI, sender, receiver *itutiltypes.TestAccount, tokenId *big.Int) (*evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(sender)
	suite.Require().NotNil(receiver)
	suite.Require().NotNil(tokenId)

	data := suite.prepareTransferNft721TokenDara(abi, sender, receiver, tokenId)
	return suite.TxSendEvmTx(suite.CurrentContext, sender, &contract, nil, data)
}

// TxTransferNft721TokenAsync is the same as TxTransferNft721Token but with Async delivery mode.
func (suite *ChainIntegrationTestSuite) TxTransferNft721TokenAsync(contract common.Address, abi abi.ABI, sender, receiver *itutiltypes.TestAccount, tokenId *big.Int) (*evmtypes.MsgEthereumTx, error) {
	suite.Require().NotNil(sender)
	suite.Require().NotNil(receiver)
	suite.Require().NotNil(tokenId)

	data := suite.prepareTransferNft721TokenDara(abi, sender, receiver, tokenId)
	return suite.TxSendEvmTxAsync(suite.CurrentContext, sender, &contract, nil, data)
}

// prepareTransferNft721TokenDara computes the call data for the "safeTransferFrom" function of the given ERC-721 contract, with given tokenId of token to transfer.
func (suite *ChainIntegrationTestSuite) prepareTransferNft721TokenDara(abi abi.ABI, sender, receiver *itutiltypes.TestAccount, tokenId *big.Int) []byte {
	suite.Require().NotNil(sender)
	suite.Require().NotNil(receiver)
	suite.Require().NotNil(tokenId)

	data, err := abi.Pack("safeTransferFrom", sender.GetEthAddress(), receiver.GetEthAddress(), tokenId)
	suite.Require().NoError(err)
	return data
}

// TxTransferNft1155Token call the "safeTransferFrom" function of the given ERC-1155 contract, transferring give amount of given NFT token from sender to receiver.
func (suite *ChainIntegrationTestSuite) TxTransferNft1155Token(contract common.Address, abi abi.ABI, sender, receiver *itutiltypes.TestAccount, tokenId *big.Int, amount uint16) (*evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	suite.Require().NotNil(sender)
	suite.Require().NotNil(receiver)
	suite.Require().NotNil(tokenId)

	data := suite.prepareTransferNft1155TokenData(abi, sender, receiver, tokenId, amount)
	return suite.TxSendEvmTx(suite.CurrentContext, sender, &contract, nil, data)
}

// TxTransferNft1155TokenAsync is the same as TxTransferNft1155Token but with Async delivery mode.
func (suite *ChainIntegrationTestSuite) TxTransferNft1155TokenAsync(contract common.Address, abi abi.ABI, sender, receiver *itutiltypes.TestAccount, tokenId *big.Int, amount uint16) (*evmtypes.MsgEthereumTx, error) {
	suite.Require().NotNil(sender)
	suite.Require().NotNil(receiver)
	suite.Require().NotNil(tokenId)

	data := suite.prepareTransferNft1155TokenData(abi, sender, receiver, tokenId, amount)
	return suite.TxSendEvmTxAsync(suite.CurrentContext, sender, &contract, nil, data)
}

// prepareTransferNft1155TokenData computes the call data for the "safeTransferFrom" function of the given ERC-1155 contract, with given tokenId and amount of token to transfer.
func (suite *ChainIntegrationTestSuite) prepareTransferNft1155TokenData(abi abi.ABI, sender, receiver *itutiltypes.TestAccount, tokenId *big.Int, amount uint16) []byte {
	suite.Require().NotNil(sender)
	suite.Require().NotNil(receiver)
	suite.Require().NotNil(tokenId)

	amt := new(big.Int).SetInt64(int64(amount))

	data, err := abi.Pack("safeTransferFrom", sender.GetEthAddress(), receiver.GetEthAddress(), tokenId, amt, []byte{})
	suite.Require().NoError(err)
	return data
}

// TxSendEvmTx builds and sends a MsgEthereumTx message based on the given call-data.
// The given sender account will be used to sign the message.
func (suite *ChainIntegrationTestSuite) TxSendEvmTx(ctx sdk.Context, sender *itutiltypes.TestAccount, to *common.Address, amount *big.Int, inputCallData []byte) (*evmtypes.MsgEthereumTx, *itutiltypes.ResponseDeliverEthTx, error) {
	msgEthereumTx := suite.prepareMsgEthereumTx(ctx, sender, to, amount, inputCallData, 0)
	resDeliverEthTx, err := suite.DeliverEthTx(sender, msgEthereumTx)
	return msgEthereumTx, resDeliverEthTx, err
}

// TxSendEvmTxAsync is the same as TxSendEvmTx but with Async delivery mode.
func (suite *ChainIntegrationTestSuite) TxSendEvmTxAsync(ctx sdk.Context, sender *itutiltypes.TestAccount, to *common.Address, amount *big.Int, inputCallData []byte) (*evmtypes.MsgEthereumTx, error) {
	msgEthereumTx := suite.prepareMsgEthereumTx(ctx, sender, to, amount, inputCallData, 0)
	err := suite.DeliverEthTxAsync(sender, msgEthereumTx)
	return msgEthereumTx, err
}

// prepareMsgEthereumTx builds a MsgEthereumTx message based on the given input data.
func (suite *ChainIntegrationTestSuite) prepareMsgEthereumTx(ctx sdk.Context, sender *itutiltypes.TestAccount, to *common.Address, amount *big.Int, inputCallData []byte, optionalGas uint64) *evmtypes.MsgEthereumTx {
	suite.Require().NotNil(sender)

	from := sender.GetEthAddress()

	var gas uint64 = 6_000_000
	if optionalGas > 0 {
		gas = optionalGas
	}

	evmTxArgs := &evmtypes.EvmTxArgs{
		ChainID:   suite.ChainApp.EvmKeeper().ChainID(),
		Nonce:     suite.ChainApp.EvmKeeper().GetNonce(ctx, from),
		GasLimit:  gas,
		GasFeeCap: suite.ChainApp.FeeMarketKeeper().GetBaseFee(ctx),
		GasTipCap: big.NewInt(1),
		To:        to,
		Amount:    amount,
		Input:     inputCallData,
		Accesses:  &ethtypes.AccessList{},
	}

	msgEthereumTx := evmtypes.NewTx(evmTxArgs)
	msgEthereumTx.From = from.String()

	return msgEthereumTx
}

// computeAmount computes the amount of token to transfer, based on the given amount and decimals.
func computeAmount(amount uint16, decimals uint8) sdkmath.Int {
	intDecimal := sdkmath.NewInt(int64(math.Pow10(int(decimals))))
	intAmount := sdkmath.NewInt(int64(amount))
	return intAmount.Mul(intDecimal)
}
