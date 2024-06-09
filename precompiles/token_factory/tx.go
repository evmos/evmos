// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package tokenfactory

import (
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	access "github.com/evmos/evmos/v18/precompiles/access_control"
	erc20types "github.com/evmos/evmos/v18/x/erc20/types"
)

const (
	MethodCreateERC20  = "createERC20"
	MethodCreate2ERC20 = "create2ERC20"
)

// CreateERC20 creates a new Bank coin with an ERC20 precompiled interface with the given parameters
// It adds the relevant metadata and role permissions to the contract owner
func (p Precompile) CreateERC20(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	name, symbol, decimals, initialSupply, err := ParseCreateErc20Args(args)
	if err != nil {
		return nil, err
	}

	precompileAddr := p.Address()
	account := p.accountKeeper.GetAccount(ctx, precompileAddr.Bytes())
	if account == nil {
		account = p.accountKeeper.NewAccountWithAddress(ctx, precompileAddr.Bytes())
	}

	address := crypto.CreateAddress(p.Address(), account.GetSequence())

	return p.createERC20(ctx, contract, stateDB, method, address, name, symbol, decimals, initialSupply, account)
}

// Create2ERC20 creates a new Bank coin with an ERC20 precompiled interface with the given parameters
// It adds the relevant metadata and role permissions to the contract owner
// It uses a salt to deterministically create the address
func (p Precompile) Create2ERC20(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	name, symbol, decimals, initialSupply, salt, err := ParseCreate2Erc20Args(args)
	if err != nil {
		return nil, err
	}

	precompileAddr := p.Address()
	account := p.accountKeeper.GetAccount(ctx, precompileAddr.Bytes())
	if account == nil {
		account = p.accountKeeper.NewAccountWithAddress(ctx, precompileAddr.Bytes())
	}

	hash := common.Hash{}
	address := crypto.CreateAddress2(p.Address(), salt, hash.Bytes())

	return p.createERC20(ctx, contract, stateDB, method, address, name, symbol, decimals, initialSupply, account)
}

// createERC20 creates a new ERC20 token with the given parameters
func (p Precompile) createERC20(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	address common.Address,
	name, symbol string,
	decimals uint8,
	initialSupply *big.Int,
	account authtypes.AccountI,
) ([]byte, error) {
	addrHex := address.String()
	denom := erc20types.CreateDenom(addrHex)

	tokenPair := erc20types.TokenPair{
		Erc20Address:  addrHex,
		Denom:         denom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_EXTERNAL,
	}

	// Save the token pair in the store
	p.erc20Keeper.SetTokenPair(ctx, tokenPair)
	p.erc20Keeper.SetDenomMap(ctx, tokenPair.Denom, tokenPair.GetID())
	p.erc20Keeper.SetERC20Map(ctx, common.HexToAddress(tokenPair.Erc20Address), tokenPair.GetID())

	// TODO: there needs to be logic here that creates different token factory variants
	erc20ACPrecompile, err := access.NewPrecompile(tokenPair, p.bankKeeper, p.authzKeeper, p.transferKeeper, p.acKeeper)
	if err != nil {
		return nil, err
	}

	if err := p.evmKeeper.AddDynamicPrecompiles(ctx, erc20ACPrecompile); err != nil {
		return nil, err
	}

	metadata := NewDenomMetaData(addrHex, denom, name, symbol, decimals)
	if err := metadata.Validate(); err != nil {
		return nil, err
	}

	p.bankKeeper.SetDenomMetaData(ctx, metadata)

	// Set the default roles for the contract owner
	// TODO: This only needs to be present on the access control token factory variant
	p.acKeeper.SetRole(ctx, address, access.RoleDefaultAdmin, contract.CallerAddress)
	p.acKeeper.SetRole(ctx, address, access.RoleMinter, contract.CallerAddress)
	p.acKeeper.SetRole(ctx, address, access.RoleBurner, contract.CallerAddress)

	// TODO: emit events RoleGranted and RoleAdminChanged

	// TODO: set owner to caller address

	if initialSupply != nil && initialSupply.Sign() > 0 {
		if err := p.bankKeeper.MintCoins(ctx, erc20types.ModuleName, sdk.Coins{{Denom: denom, Amount: math.NewIntFromBigInt(initialSupply)}}); err != nil {
			return nil, err
		}

		if err := p.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			erc20types.ModuleName,
			contract.CallerAddress.Bytes(),
			sdk.Coins{{Denom: denom, Amount: math.NewIntFromBigInt(initialSupply)}},
		); err != nil {
			return nil, err
		}

		if err := p.EmitEventMint(ctx, stateDB, contract.CallerAddress, initialSupply); err != nil {
			return nil, err
		}
	}

	nonce := account.GetSequence()
	if err := account.SetSequence(nonce + 1); err != nil {
		return nil, err
	}

	p.accountKeeper.SetAccount(ctx, account)

	if err := p.EmitCreateERC20Event(ctx, stateDB, contract.CallerAddress, address, name, symbol, decimals, initialSupply); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
