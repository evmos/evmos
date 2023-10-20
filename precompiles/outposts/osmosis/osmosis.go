// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis

import (
	"bytes"
	"embed"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
	erc20keeper "github.com/evmos/evmos/v14/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v14/x/erc20/types"
	transferkeeper "github.com/evmos/evmos/v14/x/ibc/transfer/keeper"
)

const (
	// OsmosisChannelIDMainnet is the channel ID for the Osmosis channel on Evmos mainnet.
	OsmosisChannelIDMainnet = "channel-0"
	// OsmosisChannelIDTestnet is the channel ID for the Osmosis channel on Evmos testnet.
	OsmosisChannelIDTestnet = "channel-0"

	// OsmosisOutpostAddress is the address of the Osmosis outpost precompile
	OsmosisOutpostAddress   = "0x0000000000000000000000000000000000000901"
)

const (
	// TimeoutHeight is the default value used in the IBC timeout height for
	// the client.
	DefaultTimeoutHeight = 100
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//go:embed abi.json
var f embed.FS

/// Precompile is the structure that define the Osmosis outpost precompiles extending 
/// the common Precompile type.
type Precompile struct {
	cmn.Precompile
	// IBC 
	portID             string
	channelID          string
	timeoutHeight      clienttypes.Height

	// Osmosis
	osmosisXCSContract string

	// Keepers
	bankKeeper     erc20types.BankKeeper
	transferKeeper transferkeeper.Keeper
	erc20Keeper    erc20keeper.Keeper
}

// NewPrecompile creates a new Osmosis outpost Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	portID, channelID string,
	osmosisXCSContract string,
	bankKeeper erc20types.BankKeeper,
	transferKeeper transferkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) (*Precompile, error) {
	abiBz, err := f.ReadFile("abi.json")
	if err != nil {
		return nil, err
	}

	newAbi, err := abi.JSON(bytes.NewReader(abiBz))
	if err != nil {
		return nil, err
	}

	return &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newAbi,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration, // should be configurable in the future.
		},
		portID:             portID,
		channelID:          channelID,
		timeoutHeight:      clienttypes.NewHeight(DefaultTimeoutHeight, DefaultTimeoutHeight),
		osmosisXCSContract: osmosisXCSContract,
		transferKeeper:     transferKeeper,
		bankKeeper:         bankKeeper,
		erc20Keeper:        erc20Keeper,
	}, nil
}

// Address defines the address of the Osmosis outpost precompile contract.
func (Precompile) Address() common.Address {
	return common.HexToAddress(OsmosisChannelIDTestnet)
}
