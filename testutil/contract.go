package testutil

import (
	"errors"
	"math/big"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/testutil/tx"
	evm "github.com/evmos/evmos/v11/x/evm/types"
)

// DeployContract deploys a contract with the provided private key,
// compiled contract data and constructor arguments
func DeployContract(
	ctx sdk.Context,
	evmosApp *app.Evmos,
	priv cryptotypes.PrivKey,
	queryClientEvm evm.QueryClient,
	contract evm.CompiledContract,
	constructorArgs ...interface{},
) (common.Address, error) {
	chainID := evmosApp.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := evmosApp.EvmKeeper.GetNonce(ctx, from)

	ctorArgs, err := contract.ABI.Pack("", constructorArgs...)
	if err != nil {
		return common.Address{}, err
	}

	data := append(contract.Bin, ctorArgs...) //nolint:gocritic
	gas, err := tx.GasLimit(ctx, from, data, queryClientEvm)
	if err != nil {
		return common.Address{}, err
	}

	msgEthereumTx := evm.NewTx(&evm.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		GasLimit:  gas,
		GasFeeCap: evmosApp.FeeMarketKeeper.GetBaseFee(ctx),
		GasTipCap: big.NewInt(1),
		Input:     data,
		Accesses:  &ethtypes.AccessList{},
	})
	msgEthereumTx.From = from.String()

	if _, err = DeliverEthTx(evmosApp, priv, msgEthereumTx); err != nil {
		return common.Address{}, err
	}

	return getContractAddr(ctx, evmosApp, from, nonce)
}

// DeployContractWithFactory deploys a contract using a contract factory
// with the provided factoryAddress
func DeployContractWithFactory(
	ctx sdk.Context,
	evmosApp *app.Evmos,
	priv cryptotypes.PrivKey,
	factoryAddress common.Address,
	queryClientEvm evm.QueryClient,
) (common.Address, abci.ResponseDeliverTx, error) {
	chainID := evmosApp.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	factoryNonce := evmosApp.EvmKeeper.GetNonce(ctx, factoryAddress)
	nonce := evmosApp.EvmKeeper.GetNonce(ctx, from)

	msgEthereumTx := evm.NewTx(&evm.EvmTxArgs{
		ChainID:  chainID,
		Nonce:    nonce,
		To:       &factoryAddress,
		GasLimit: uint64(100000),
		GasPrice: big.NewInt(1000000000),
	})
	msgEthereumTx.From = from.String()

	res, err := DeliverEthTx(evmosApp, priv, msgEthereumTx)
	if err != nil {
		return common.Address{}, abci.ResponseDeliverTx{}, err
	}

	addr, err := getContractAddr(ctx, evmosApp, factoryAddress, factoryNonce)
	return addr, res, err
}

// getContractAddr calculates the contract address based on the deployer's address and its nonce
// Then, checks if the account exists and has the 'code' field populated
func getContractAddr(ctx sdk.Context, evmosApp *app.Evmos, from common.Address, nonce uint64) (common.Address, error) {
	contractAddress := crypto.CreateAddress(from, nonce)
	acc := evmosApp.EvmKeeper.GetAccountWithoutBalance(ctx, contractAddress)
	if acc == nil {
		return common.Address{}, errors.New("an error occurred when deploying the contract. GetAccountWithoutBalance using contract's account returned nil")
	}
	if !acc.IsContract() {
		return common.Address{}, errors.New("an error occurred when deploying the contract. Contract's account does not have the contract code")
	}
	return contractAddress, nil
}
