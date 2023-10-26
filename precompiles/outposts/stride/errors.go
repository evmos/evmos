// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride

const (
	// ErrTokenPairNotFound is the error returned when a token pair is not found
	// #nosec G101
	ErrTokenPairNotFound = "token pair not found for %s"
	// ErrUnsupportedToken is the error returned when a token is not supported
	ErrUnsupportedToken = "unsupported token %s. The only supported token contract for Stride Outpost v1 is %s"
	// ErrRedeemStakeEmptyIBCReceiver is the error returned when the receiver is empty
	ErrRedeemStakeEmptyIBCReceiver = "IBCReceiver cannot be empty when action is RedeemStake"
	// ErrEmptyReceiver is the error returned when the receiver is empty
	ErrEmptyReceiver = "receiver cannot be empty"
	// ErrEmptyAutopilotAction is the error returned when the autopilot action is empty
	ErrEmptyAutopilotAction = "autopilot action cannot be empty"
)
