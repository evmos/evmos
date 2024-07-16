// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmostypes "github.com/evmos/evmos/v18/types"
	"github.com/evmos/evmos/v18/utils"
)

// NewTokenPairSTRv2 creates a new TokenPair instance in the context of the
// Single Token Representation v2.
//
// It derives the ERC-20 address from the hex suffix of the IBC denomination
// (e.g. ibc/DF63978F803A2E27CA5CC9B7631654CCF0BBC788B3B7F0A10200508E37C70992).
func NewTokenPairSTRv2(denom string) (TokenPair, error) {
	address, err := utils.GetIBCDenomAddress(denom)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		Erc20Address:  address.String(),
		Denom:         denom,
		Enabled:       true,
		ContractOwner: OWNER_MODULE,
	}, nil
}

// NewTokenPair returns an instance of TokenPair
func NewTokenPair(erc20Address common.Address, denom string, contractOwner Owner) TokenPair {
	return TokenPair{
		Erc20Address:  erc20Address.String(),
		Denom:         denom,
		Enabled:       true,
		ContractOwner: contractOwner,
	}
}

// GetID returns the SHA256 hash of the ERC20 address and denomination
func (tp TokenPair) GetID() []byte {
	id := tp.Erc20Address + "|" + tp.Denom
	return tmhash.Sum([]byte(id))
}

// GetErc20Contract casts the hex string address of the ERC20 to common.Address
func (tp TokenPair) GetERC20Contract() common.Address {
	return common.HexToAddress(tp.Erc20Address)
}

// Validate performs a stateless validation of a TokenPair
func (tp TokenPair) Validate() error {
	if err := sdk.ValidateDenom(tp.Denom); err != nil {
		return err
	}

	return evmostypes.ValidateAddress(tp.Erc20Address)
}

// IsNativeCoin returns true if the owner of the ERC20 contract is the
// erc20 module account
func (tp TokenPair) IsNativeCoin() bool {
	return tp.ContractOwner == OWNER_MODULE
}

// IsNativeERC20 returns true if the owner of the ERC20 contract is an EOA.
func (tp TokenPair) IsNativeERC20() bool {
	return tp.ContractOwner == OWNER_EXTERNAL
}
