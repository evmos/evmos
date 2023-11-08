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
	feeCollectorName string
	bankKeeper       bankkeeper.Keeper
}

func (h HandlerOptions) Validate() error {
	if h.feeCollectorName == "" {
		return errors.New("fee collector name cannot be empty")
	}

	if h.bankKeeper == nil {
		return errors.New("bank keeper cannot be nil")
	}

	return nil
}

// NewPostHandler returns an empty PostHandler chain.
func NewPostHandler(ho HandlerOptions) sdk.PostHandler {
	postDecorators := []sdk.PostDecorator{
		NewBurnDecorator(ho.feeCollectorName, ho.bankKeeper),
	}

	return sdk.ChainPostDecorators(postDecorators...)
}
