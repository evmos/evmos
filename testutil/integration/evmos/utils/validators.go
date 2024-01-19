package utils

import (
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ConvertValAddressToHex(valAddress string) common.Address {
	coinbaseAddressBytes := sdk.ConsAddress(valAddress).Bytes()
	return common.BytesToAddress(coinbaseAddressBytes)
}
