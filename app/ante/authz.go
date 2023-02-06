// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// AuthzLimiterDecorator blocks certain msg types from being granted or executed within authz.
type AuthzLimiterDecorator struct {
	// disabledMsgTypes is the type urls of the msgs to block.
	disabledMsgTypes []string
}

// NewAuthzLimiterDecorator creates a decorator to block certain msg types from being granted or executed within authz.
func NewAuthzLimiterDecorator(disabledMsgTypes ...string) AuthzLimiterDecorator {
	return AuthzLimiterDecorator{
		disabledMsgTypes: disabledMsgTypes,
	}
}

func (ald AuthzLimiterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	err = ald.checkForDisabledMsg(tx.GetMsgs(), true)
	if err != nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%v", err)
	}
	return next(ctx, tx, simulate)
}

// checkForDisabledMsg iterates through the msgs and returns an error if it finds any unauthorized msgs.
//
// When searchOnlyInAuthzMsgs is enabled, only authz MsgGrant and MsgExec are blocked, if they contain unauthorized msg types.
// Otherwise any msg matching the disabled types are blocked, regardless of being in an authz msg or not.
//
// This method is recursive as MsgExec's can wrap other MsgExecs.
func (ald AuthzLimiterDecorator) checkForDisabledMsg(msgs []sdk.Msg, searchOnlyInAuthzMsgs bool) error {
	for _, msg := range msgs {
		typeURL := sdk.MsgTypeURL(msg)
		switch {
		case !searchOnlyInAuthzMsgs && ald.isDisabled(typeURL):
			return fmt.Errorf("found disabled msg type: %s", typeURL)

		case typeURL == sdk.MsgTypeURL(&authz.MsgGrant{}):
			m, ok := msg.(*authz.MsgGrant)
			if !ok {
				panic("unexpected msg type")
			}
			authorization, err := m.GetAuthorization()
			if err != nil {
				return err
			}
			if ald.isDisabled(authorization.MsgTypeURL()) {
				return fmt.Errorf("found disabled msg type in MsgGrant: %s", authorization.MsgTypeURL())
			}

		case typeURL == sdk.MsgTypeURL(&authz.MsgExec{}):
			m, ok := msg.(*authz.MsgExec)
			if !ok {
				panic("unexpected msg type")
			}
			innerMsgs, err := m.GetMessages()
			if err != nil {
				return err
			}
			if err := ald.checkForDisabledMsg(innerMsgs, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ald AuthzLimiterDecorator) isDisabled(msgTypeURL string) bool {
	for _, disabledType := range ald.disabledMsgTypes {
		if msgTypeURL == disabledType {
			return true
		}
	}
	return false
}
