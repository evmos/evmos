package keeper

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/evmos/v18/server/config"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

// NOTE: These selectors were generated using the `functionName.selector` syntax in Solidity.
// You can read more here in the 'How Interfaces are Identified' section - https://eips.ethereum.org/EIPS/eip-165
var (
	OnSendPacketInterfaceID    = [4]byte{0x1e, 0xe6, 0x8e, 0x81} // 0x1ee68e81
	OnAckPacketInterfaceID     = [4]byte{0x50, 0x81, 0x24, 0x4a} // 0x5081244a
	OnRecvPacketInterfaceID    = [4]byte{0x50, 0x81, 0x24, 0x4a} // 0x5081244a
	OnTimeoutPacketInterfaceID = [4]byte{0xdd, 0x14, 0xd3, 0xbd} // 0xdd14d3bd
)

// DetectInterface checks if the contract at the given address supports the given interfaceID.
// It does this by calling the `supportsInterface` function on the contract.
func (k Keeper) DetectInterface(cachedCtx sdk.Context, interfaceID [4]byte, contractHex common.Address) error {
	input, err := k.ABI.Pack(SupportsInterfaceQuery, interfaceID)
	if err != nil {
		return err
	}

	//packetSender := common.HexToAddress(packetSenderAddress)
	callArgs := evmtypes.TransactionArgs{
		To:   &contractHex,
		Data: (*hexutil.Bytes)(&input),
	}

	bz, err := json.Marshal(&callArgs)
	if err != nil {
		return err
	}

	callReq := evmtypes.EthCallRequest{
		Args:            bz,
		GasCap:          config.DefaultGasCap,
		ProposerAddress: cachedCtx.BlockHeader().ProposerAddress,
		ChainId:         k.evmKeeper.ChainID().Int64(),
	}

	res, err := k.evmKeeper.EthCall(cachedCtx, &callReq)
	if err != nil {
		return err
	}

	unpacked, err := k.ABI.Unpack(SupportsInterfaceQuery, res.Ret)
	if err != nil {
		return err
	}

	if unpacked[0] != true {
		return fmt.Errorf("contract does not support interface %x", interfaceID)
	}

	return err
}
