package testutil

import (
	"encoding/json"
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/evmos/evmos/v11/app"
	"github.com/evmos/evmos/v11/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v11/server/config"
	evm "github.com/evmos/evmos/v11/x/evm/types"
)

// DeployContract deploys a contract with the provided private key,
// compiled contract data and constructor arguments
func DeployContract(
	ctx sdk.Context,
	app *app.Evmos,
	priv *ethsecp256k1.PrivKey,
	queryClientEvm evm.QueryClient,
	contract evm.CompiledContract,
	constructorArgs ...interface{},
) (common.Address, error) {
	chainID := app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := app.EvmKeeper.GetNonce(ctx, from)

	ctorArgs, err := contract.ABI.Pack("", constructorArgs...)
	if err != nil {
		return common.Address{}, err
	}

	data := append(contract.Bin, ctorArgs...) //nolint:gocritic
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &from,
		Data: (*hexutil.Bytes)(&data),
	})
	if err != nil {
		return common.Address{}, err
	}

	goCtx := sdk.WrapSDKContext(ctx)
	res, err := queryClientEvm.EstimateGas(goCtx, &evm.EthCallRequest{
		Args:   args,
		GasCap: config.DefaultGasCap,
	})
	if err != nil {
		return common.Address{}, err
	}

	msgEthereumTx := evm.NewTxContract(
		chainID,
		nonce,
		nil,     // amount
		res.Gas, // gasLimit
		nil,     // gasPrice
		app.FeeMarketKeeper.GetBaseFee(ctx),
		big.NewInt(1),
		data,                   // input
		&ethtypes.AccessList{}, // accesses
	)
	msgEthereumTx.From = from.String()

	if _, err = DeliverEthTx(app, priv, msgEthereumTx); err != nil {
		return common.Address{}, err
	}

	contractAddress := crypto.CreateAddress(from, nonce)
	acc := app.EvmKeeper.GetAccountWithoutBalance(ctx, contractAddress)
	if acc == nil {
		return common.Address{}, errors.New("an error occurred when creating the contract. GetAccountWithoutBalance using contract's account returned nil")
	}
	if !acc.IsContract() {
		return common.Address{}, errors.New("an error occurred when creating the contract. Contract's account does not have the contract code")
	}

	return contractAddress, nil
}
