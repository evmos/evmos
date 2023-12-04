// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package post

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// HandlerOptions are the options required for constructing a PostHandler.
type HandlerOptions struct {
	FeeCollectorName string
	BankKeeper       bankkeeper.Keeper
}

func (h HandlerOptions) Validate() error {
	if h.FeeCollectorName == "" {
		return errors.New("fee collector name cannot be empty")
	}

	if h.BankKeeper == nil {
		return errors.New("bank keeper cannot be nil")
	}

	return nil
}

// NewPostHandler returns a new PostHandler decorators chain.
func NewPostHandler(ho HandlerOptions) sdk.PostHandler {
	postDecorators := []sdk.PostDecorator{
		NewBurnDecorator(ho.FeeCollectorName, ho.BankKeeper),
	}

	return sdk.ChainPostDecorators(postDecorators...)
}
