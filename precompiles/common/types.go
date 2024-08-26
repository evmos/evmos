// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package common

import (
	"time"

	"github.com/evmos/evmos/v19/contracts/types"
	evmosutils "github.com/evmos/evmos/v19/utils"
)

var (
	// TrueValue is the byte array representing a true value in solidity.
	TrueValue = []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}
	// DefaultExpirationDuration is the default duration for an authorization to expire.
	DefaultExpirationDuration = time.Hour * 24 * 365
	// DefaultChainID is the standard chain id used for testing purposes
	DefaultChainID = evmosutils.MainnetChainID + "-1"
)

// ICS20Allocation defines the spend limit for a particular port and channel.
// We need this to be able to unpack to big.Int instead of math.Int.
type ICS20Allocation struct {
	SourcePort        string
	SourceChannel     string
	SpendLimit        []types.Coin
	AllowList         []string
	AllowedPacketData []string
}
