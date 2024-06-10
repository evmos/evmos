// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package contracts

import (
	"math/big"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// CallArgs is a struct to define all relevant data to call a smart contract.
type CallArgs struct {
	// Amount is the msg.value sent with the transaction.
	Amount *big.Int
	// AccessList is the access list to use for the transaction. If not empty, the transaction will be an EIP-2930 transaction (AccessListTx).
	AccessList *ethtypes.AccessList
	// ContractAddr is the address of the contract to call.
	ContractAddr common.Address
	// ContractABI is the ABI of the contract to call.
	ContractABI abi.ABI
	// MethodName is the name of the method to call.
	MethodName string
	// Nonce is the nonce to use for the transaction.
	Nonce *big.Int
	// GasLimit to use for the transaction
	GasLimit uint64
	// GasPrice is the gas price to use. If left empty, the base fee for the current block will be used.
	GasPrice *big.Int
	// GasFeeCap is the gas fee cap to use. If not empty, the transaction will be an EIP-1559 transaction (DynamicFeeTx).
	GasFeeCap *big.Int
	// GasTipCap is the gas tip cap to use. If not empty, the transaction will be an EIP-1559 transaction (DynamicFeeTx).
	GasTipCap *big.Int
	// PrivKey is the private key to be used for the transaction.
	PrivKey cryptotypes.PrivKey
	// Args are the arguments to pass to the method.
	Args []interface{}
}

// WithAddress returns the CallArgs with the given address.
func (c CallArgs) WithAddress(addr common.Address) CallArgs {
	c.ContractAddr = addr
	return c
}

// WithABI returns the CallArgs with the given contract ABI.
func (c CallArgs) WithABI(abi abi.ABI) CallArgs {
	c.ContractABI = abi
	return c
}

// WithMethodName returns the CallArgs with the given method name.
func (c CallArgs) WithMethodName(methodName string) CallArgs {
	c.MethodName = methodName
	return c
}

// WithNonce returns the CallArgs with the given nonce.
func (c CallArgs) WithNonce(nonce *big.Int) CallArgs {
	c.Nonce = nonce
	return c
}

// WithGasLimit returns the CallArgs with the given gas limit.
func (c CallArgs) WithGasLimit(gasLimit uint64) CallArgs {
	c.GasLimit = gasLimit
	return c
}

// WithPrivKey returns the CallArgs with the given private key.
func (c CallArgs) WithPrivKey(privKey cryptotypes.PrivKey) CallArgs {
	c.PrivKey = privKey
	return c
}

// WithArgs populates the CallArgs struct's Args field with the given list of arguments.
// These are the arguments that will be packed into the contract call input.
func (c CallArgs) WithArgs(args ...interface{}) CallArgs {
	c.Args = append([]interface{}{}, args...)
	return c
}

// WithAmount populates the CallArgs struct's Amount field with the given amount.
// This is the amount of Evmos that will be sent with the contract call.
func (c CallArgs) WithAmount(amount *big.Int) CallArgs {
	c.Amount = amount
	return c
}

// WithGasPrice returns the CallArgs with the given gas price.
func (c CallArgs) WithGasPrice(gasPrice *big.Int) CallArgs {
	c.GasPrice = gasPrice
	return c
}

// WithGasFeeCap returns the CallArgs with the given gas fee cap.
func (c CallArgs) WithGasFeeCap(gasFeeCap *big.Int) CallArgs {
	c.GasFeeCap = gasFeeCap
	return c
}

// WithGasTipCap returns the CallArgs with the given gas tip cap.
func (c CallArgs) WithGasTipCap(gasTipCap *big.Int) CallArgs {
	c.GasTipCap = gasTipCap
	return c
}

// WithAccessList returns the CallArgs with the given access list.
func (c CallArgs) WithAccessList(accessList *ethtypes.AccessList) CallArgs {
	c.AccessList = accessList
	return c
}
