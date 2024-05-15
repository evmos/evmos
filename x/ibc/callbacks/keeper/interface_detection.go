package keeper

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/evmos/v17/server/config"
	evmtypes "github.com/evmos/evmos/v17/x/evm/types"
)

// NOTE: These selectors were generated using the `functionName.selector` syntax in Solidity.
// You can read more here in the 'How Interfaces are Identified' section - https://eips.ethereum.org/EIPS/eip-165
var (
	OnAckPacketInterfaceID     = [4]byte{0x40, 0xe6, 0xba, 0xfd} // 0x40e6bafd
	OnSendPacketInterfaceID    = [4]byte{0x1e, 0xe6, 0x8e, 0x81} // 0x1ee68e81
	OnRecvPacketInterfaceID    = [4]byte{0xeb, 0x8f, 0xcd, 0x41} // 0xeb8fcd41
	OnTimeoutPacketInterfaceID = [4]byte{0x02, 0x08, 0x72, 0xcd} // 0x020872cd
)

// DetectInterface checks if the contract at the given address supports the given interfaceID.
// It does this by calling the `supportsInterface` function on the contract.
func (k Keeper) DetectInterface(cachedCtx sdk.Context, interfaceID [4]byte, packetSenderAddress string, contractHex common.Address) error {
	input, err := k.ABI.Pack(SupportsInterfaceQuery, interfaceID)
	if err != nil {
		fmt.Println("The error in packing SupportInterfaceQuery is", err)
		return err
	}

	packetSender := common.HexToAddress(packetSenderAddress)
	callArgs := evmtypes.TransactionArgs{
		From: &packetSender,
		To:   &contractHex,
		Data: (*hexutil.Bytes)(&input),
	}

	bz, err := json.Marshal(&callArgs)
	if err != nil {
		fmt.Println("The error in marshalling is", err)
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
		fmt.Println("The error in ETH CALL is", err)
		return err
	}

	fmt.Println("The result in ETH CALL is", res.Ret, res.VmError, res.Logs, res.Hash, res.Failed())

	unpacked, err := k.ABI.Unpack(SupportsInterfaceQuery, res.Ret)
	if err != nil {
		fmt.Println("The error in unpacking is", err)
		return err
	}

	if unpacked[0] != true {
		return fmt.Errorf("contract does not support interface %x", interfaceID)
	}

	fmt.Println("The unpacked is", unpacked)
	return err
}
