package types

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func GetBaseFee(
	ctx context.Context,
	client evmtypes.QueryClient,
) (*math.Int, error) {
	req := &evmtypes.QueryBaseFeeRequest{}
	res, err := client.BaseFee(ctx, req)
	if err != nil {
		return nil, err
	}

	if res == nil || res.BaseFee == nil {
		return nil, fmt.Errorf("base fee response is nil")
	}

	return res.BaseFee, nil
}

func GetProposerAddres(
	ctx context.Context,
	client evmtypes.QueryClient,
	proposer sdk.ConsAddress,
) (common.Address, error) {
	req := &evmtypes.QueryValidatorAccountRequest{
		ConsAddress: proposer.String(),
	}

	res, err := client.ValidatorAccount(ctx, req)
	if err != nil {
		return common.Address{}, err
	}

	validatorAccAddr, err := sdk.AccAddressFromBech32(res.AccountAddress)
	if err != nil {
		return common.Address{}, err
	}

	return common.BytesToAddress(validatorAccAddr), nil
}

func GetBlockGasUsed(
	txs [][]byte,
	txDecoder sdk.TxDecoder,
) (uint64, error) {
	gasUsed := uint64(0)
	for _, txBz := range txs {
		tx, err := txDecoder(txBz)
		if err != nil {
			return 0, err
		}

		for _, msg := range tx.GetMsgs() {
			_, txData, _, err := evmtypes.UnpackEthMsg(msg)
			if err != nil {
				continue
			}

			// TODO: this is the current txs not the past tx result block
			gasUsed += txData.GetGas()
		}
	}

	return gasUsed, nil
}
