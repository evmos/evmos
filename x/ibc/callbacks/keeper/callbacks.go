package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v17/crypto/ethsecp256k1"
	evmtypes "github.com/evmos/evmos/v17/x/evm/types"
	"math/big"
)

const (
	SupportsInterfaceQuery         = "supportsInterface"
	IBCSendPacketMethod            = "onSendPacket"
	IBCAcknowledgementPacketMethod = "onAcknowledgementPacket"
	IBCTimeoutPacketMethod         = "onTimeoutPacket"
	IBCReceivePacketMethod         = "onReceivePacket"
)

type ICS20Packet struct {
	SourcePort       string                                `abi:"sourcePort"`
	SourceChannel    string                                `abi:"sourceChannel"`
	Data             transfertypes.FungibleTokenPacketData `abi:"data"`
	TimeoutHeight    clienttypes.Height                    `abi:"timeoutHeight"`
	TimeoutTimestamp uint64                                `abi:"timeoutTimestamp"`
}

func (k Keeper) IBCSendPacketCallback(cachedCtx sdk.Context, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, packetData []byte, contractAddress, packetSenderAddress string) error {
	cachedCtx.Logger().Info("IBCSendPacketCallback, logger")
	fmt.Println("IBCSendPacketCallback, test", sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetData, contractAddress, packetSenderAddress)
	contractHex := common.HexToAddress(contractAddress)

	if err := k.DetectInterface(cachedCtx, OnSendPacketInterfaceID, packetSenderAddress, contractHex); err != nil {
		return err
	}

	ics20Packet, err := k.DecodeTransferPacketData(packetData)
	if err != nil {
		return err
	}

	packet := ICS20Packet{
		SourcePort:       sourcePort,
		SourceChannel:    sourceChannel,
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: timeoutTimestamp,
		Data:             ics20Packet,
	}

	fmt.Println("The ics20 packet is", ics20Packet)

	newInput, err := k.ABI.Pack(IBCSendPacketMethod, packet, common.HexToAddress(packetSenderAddress))
	if err != nil {
		fmt.Println("The packing error in IBCSendPacketMethod is", err)
		return err
	}

	privkey, _ := ethsecp256k1.GenerateKey()
	key, err := privkey.ToECDSA()
	addr := crypto.PubkeyToAddress(key.PublicKey)

	chainId := k.evmKeeper.ChainID()
	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:  chainId,
		Nonce:    1,
		GasLimit: cachedCtx.GasMeter().Limit(),
		Input:    newInput,
		To:       &contractHex,
		Accesses: &ethtypes.AccessList{},
	}

	params := k.evmKeeper.GetParams(cachedCtx)
	cfg := params.GetChainConfig()
	ethCfg := cfg.EthereumConfig(chainId)

	ethSigner := ethtypes.MakeSigner(ethCfg, big.NewInt(cachedCtx.BlockHeight()))
	msgEthTx := evmtypes.NewTx(ethTxParams)
	msgEthTx.From = addr.String()
	if err := msgEthTx.Sign(ethSigner, NewSigner(privkey)); err != nil {
		fmt.Println("The error signing is", err)
		return err
	}

	txResponse, err := k.evmKeeper.EthereumTx(cachedCtx, msgEthTx)
	if err != nil {
		fmt.Println("The error tx is", err)
		return err
	}
	fmt.Println(txResponse)
	fmt.Println("IBCSendPacketCallback, test", sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetData, contractAddress, packetSenderAddress)
	return nil
}

func (k Keeper) IBCOnAcknowledgementPacketCallback(cachedCtx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress, contractAddress, packetSenderAddress string) error {
	fmt.Println("IBCOnAcknowledgementPacketCallback")
	return nil
}

func (k Keeper) IBCOnTimeoutPacketCallback(cachedCtx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress, contractAddress, packetSenderAddress string) error {
	fmt.Println("IBCOnTimeoutPacketCallback")
	return nil
}

func (k Keeper) IBCReceivePacketCallback(cachedCtx sdk.Context, packet exported.PacketI, ack exported.Acknowledgement, contractAddress string) error {
	fmt.Println("IBCReceivePacketCallback")
	return nil
}
