package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	IBCSendPacketMethod            = "onSendPacket"
	IBCAcknowledgementPacketMethod = "onAcknowledgementPacket"
	IBCTimeoutPacketMethod         = "onTimeoutPacket"
	IBCReceivePacketMethod         = "onReceivePacket"
)

func (k Keeper) IBCSendPacketCallback(cachedCtx sdk.Context, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, packetData []byte, contractAddress, packetSenderAddress string) error {
	cachedCtx.Logger().Info("IBCSendPacketCallback, logger")
	chainId := k.evmKeeper.ChainID()
	contractHex := common.HexToAddress(contractAddress)
	//packet := channeltypes.Packet{
	//	Sequence:           0,
	//	SourcePort:         sourcePort,
	//	SourceChannel:      sourceChannel,
	//	DestinationPort:    "",
	//	DestinationChannel: "",
	//	Data:               packetData,
	//	TimeoutHeight:      timeoutHeight,
	//	TimeoutTimestamp:   timeoutTimestamp,
	//}
	//args := []interface{}{packetSenderAddress}
	input, err := k.ABI.Pack(IBCSendPacketMethod, common.HexToAddress(packetSenderAddress))
	if err != nil {
		//fmt.Println("The packing error is", err)
		return err
	}
	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID: chainId,
		Nonce:   1,

		GasLimit: cachedCtx.GasMeter().Limit(),
		Input:    input,
		To:       &contractHex,
		Accesses: &ethtypes.AccessList{},
	}
	privkey, _ := ethsecp256k1.GenerateKey()
	key, err := privkey.ToECDSA()
	addr := crypto.PubkeyToAddress(key.PublicKey)

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
