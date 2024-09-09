// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
)

// ERC20RegistrationData is the necessary data to provide in order to register an ERC20 token.
type ERC20RegistrationData struct {
	// Addresses are the addresses of the ERC20 tokens.
	Addresses []string
	// ProposerPriv is the private key used to sign the proposal and voting transactions.
	ProposerPriv cryptotypes.PrivKey
}

// ValidateBasic does stateless validation of the data for the ERC20 registration.
func (ed ERC20RegistrationData) ValidateBasic() error {
	emptyAddr := common.Address{}

	if len(ed.Addresses) == 0 {
		return fmt.Errorf("addresses cannot be empty")
	}

	for _, a := range ed.Addresses {
		if ok := common.IsHexAddress(a); !ok {
			return fmt.Errorf("invalid address %s", a)
		}
		hexAddr := common.HexToAddress(a)
		if hexAddr.Hex() == emptyAddr.Hex() {
			return fmt.Errorf("address cannot be empty")
		}
	}

	if ed.ProposerPriv == nil {
		return fmt.Errorf("proposer private key cannot be nil")
	}

	return nil
}

// RegisterERC20 is a helper function to register ERC20 token through
// submitting a governance proposal and having it pass.
// It returns the registered token pair.
func RegisterERC20(tf factory.TxFactory, network network.Network, data ERC20RegistrationData) (res []erc20types.TokenPair, err error) {
	err = data.ValidateBasic()
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to validate erc20 registration data")
	}

	proposal := erc20types.MsgRegisterERC20{
		Authority:      authtypes.NewModuleAddress("gov").String(),
		Erc20Addresses: data.Addresses,
	}

	// Submit the proposal
	proposalID, err := SubmitProposal(tf, network, data.ProposerPriv, fmt.Sprintf("Register %d Token", len(data.Addresses)), &proposal)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to submit proposal")
	}

	err = network.NextBlock()
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to commit block after proposal")
	}

	// vote 'yes' and wait till proposal passes
	err = ApproveProposal(tf, network, data.ProposerPriv, proposalID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to approve proposal")
	}

	// Check if token pair is registered
	eq := network.GetERC20Client()
	for _, a := range data.Addresses {
		tokenPairRes, err := eq.TokenPair(network.GetContext(), &erc20types.QueryTokenPairRequest{Token: a})
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to query token pair")
		}
		res = append(res, tokenPairRes.TokenPair)
	}

	return res, nil
}

// ToggleTokenConversion is a helper function to toggle an ERC20 token pair conversion through
// submitting a governance proposal and having it pass.
func ToggleTokenConversion(tf factory.TxFactory, network network.Network, privKey cryptotypes.PrivKey, token string) error {
	proposal := erc20types.MsgToggleConversion{
		Authority: authtypes.NewModuleAddress("gov").String(),
		Token:     token,
	}

	// Submit the proposal
	proposalID, err := SubmitProposal(tf, network, privKey, fmt.Sprintf("Toggle %s Token", token), &proposal)
	if err != nil {
		return errorsmod.Wrap(err, "failed to submit proposal")
	}

	err = network.NextBlock()
	if err != nil {
		return errorsmod.Wrap(err, "failed to commit block after proposal")
	}

	return ApproveProposal(tf, network, privKey, proposalID)
}
