package evmv1

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	protov2 "google.golang.org/protobuf/proto"
)

// getSender extracts the sender address from the signature values using the latest signer for the given chainID.
func getSender(txData TxDataV2) (common.Address, error) {
	signer := ethtypes.LatestSignerForChainID(txData.GetChainID())
	from, err := signer.Sender(ethtypes.NewTx(txData.AsEthereumData()))
	if err != nil {
		return common.Address{}, err
	}
	return from, nil
}

func GetSigners(msg protov2.Message) ([][]byte, error) {
	msgEthTx, ok := msg.(*MsgEthereumTx)
	if !ok {
		return nil, fmt.Errorf("invalid type, expected MsgEthereumTx and got %T", msg)
	}
	var (
		txData TxDataV2
		err    error
	)

	// msgEthTx.Data is a message (DynamicFeeTx, LegacyTx or AccessListTx)
	switch msgEthTx.Data.TypeUrl {
	case "/ethermint.evm.v1.DynamicFeeTx":
		txData = &DynamicFeeTx{}
		err = msgEthTx.Data.UnmarshalTo(txData)
	case "/ethermint.evm.v1.LegacyTx":
		data := LegacyTx{}
		err = msgEthTx.Data.UnmarshalTo(&data)
	case "/ethermint.evm.v1.AccessListTx":
		data := AccessListTx{}
		err = msgEthTx.Data.UnmarshalTo(&data)
	default:
		return nil, fmt.Errorf("invalid TypeUrl %s", msgEthTx.Data.TypeUrl)
	}

	if err != nil {
		return nil, err
	}

	sender, err := getSender(txData)
	if err != nil {
		return nil, err
	}

	return [][]byte{sender.Bytes()}, nil
}
