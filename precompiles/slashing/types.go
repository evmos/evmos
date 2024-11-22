// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package slashing

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
)

// SigningInfo represents the signing info for a validator
type SigningInfo struct {
	ValidatorAddress    common.Address `abi:"validatorAddress"`
	StartHeight         uint64         `abi:"startHeight"`
	IndexOffset         uint64         `abi:"indexOffset"`
	JailedUntil         uint64         `abi:"jailedUntil"`
	Tombstoned          bool           `abi:"tombstoned"`
	MissedBlocksCounter uint64         `abi:"missedBlocksCounter"`
}

// SigningInfoOutput represents the output of the signing info query
type SigningInfoOutput struct {
	SigningInfo SigningInfo
}

// SigningInfosOutput represents the output of the signing infos query
type SigningInfosOutput struct {
	SigningInfos []SigningInfo      `abi:"signingInfos"`
	PageResponse query.PageResponse `abi:"pageResponse"`
}

// SigningInfosInput represents the input for the signing infos query
type SigningInfosInput struct {
	Pagination query.PageRequest `abi:"pagination"`
}

// ParseSigningInfoArgs parses the arguments for the signing info query
func ParseSigningInfoArgs(args []interface{}) (*slashingtypes.QuerySigningInfoRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	hexAddr, ok := args[0].(common.Address)
	if !ok || hexAddr == (common.Address{}) {
		return nil, fmt.Errorf("invalid consensus address")
	}

	return &slashingtypes.QuerySigningInfoRequest{
		ConsAddress: types.ConsAddress(hexAddr.Bytes()).String(),
	}, nil
}

// ParseSigningInfosArgs parses the arguments for the signing infos query
func ParseSigningInfosArgs(method *abi.Method, args []interface{}) (*slashingtypes.QuerySigningInfosRequest, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	var input SigningInfosInput
	if err := method.Inputs.Copy(&input, args); err != nil {
		return nil, fmt.Errorf("error while unpacking args to SigningInfosInput: %s", err)
	}

	return &slashingtypes.QuerySigningInfosRequest{
		Pagination: &input.Pagination,
	}, nil
}

func (sio *SigningInfoOutput) FromResponse(res *slashingtypes.QuerySigningInfoResponse) *SigningInfoOutput {
	sio.SigningInfo = SigningInfo{
		ValidatorAddress:    common.BytesToAddress([]byte(res.ValSigningInfo.Address)),
		StartHeight:         uint64(res.ValSigningInfo.StartHeight),        //nolint:gosec // G115
		IndexOffset:         uint64(res.ValSigningInfo.IndexOffset),        //nolint:gosec // G115
		JailedUntil:         uint64(res.ValSigningInfo.JailedUntil.Unix()), //nolint:gosec // G115
		Tombstoned:          res.ValSigningInfo.Tombstoned,
		MissedBlocksCounter: uint64(res.ValSigningInfo.MissedBlocksCounter), //nolint:gosec // G115
	}
	return sio
}

func (sio *SigningInfosOutput) FromResponse(res *slashingtypes.QuerySigningInfosResponse) *SigningInfosOutput {
	sio.SigningInfos = make([]SigningInfo, len(res.Info))
	for i, info := range res.Info {
		sio.SigningInfos[i] = SigningInfo{
			ValidatorAddress:    common.BytesToAddress([]byte(info.Address)),
			StartHeight:         uint64(info.StartHeight),        //nolint:gosec // G115
			IndexOffset:         uint64(info.IndexOffset),        //nolint:gosec // G115
			JailedUntil:         uint64(info.JailedUntil.Unix()), //nolint:gosec // G115
			Tombstoned:          info.Tombstoned,
			MissedBlocksCounter: uint64(info.MissedBlocksCounter), //nolint:gosec // G115
		}
	}
	if res.Pagination != nil {
		sio.PageResponse = query.PageResponse{
			NextKey: res.Pagination.NextKey,
			Total:   res.Pagination.Total,
		}
	}
	return sio
}

// ValidatorUnjailed defines the data structure for the ValidatorUnjailed event.
type ValidatorUnjailed struct {
	Validator common.Address
}

// Params defines the parameters for the slashing module
type Params struct {
	SignedBlocksWindow      uint64 `abi:"signedBlocksWindow"`
	MinSignedPerWindow      string `abi:"minSignedPerWindow"`
	DowntimeJailDuration    uint64 `abi:"downtimeJailDuration"`
	SlashFractionDoubleSign string `abi:"slashFractionDoubleSign"`
	SlashFractionDowntime   string `abi:"slashFractionDowntime"`
}

// ParamsOutput represents the output of the params query
type ParamsOutput struct {
	Params Params
}

func (po *ParamsOutput) FromResponse(res *slashingtypes.QueryParamsResponse) *ParamsOutput {
	po.Params = Params{
		SignedBlocksWindow:      uint64(res.Params.SignedBlocksWindow), //nolint:gosec // G115
		MinSignedPerWindow:      res.Params.MinSignedPerWindow.String(),
		DowntimeJailDuration:    uint64(res.Params.DowntimeJailDuration.Seconds()),
		SlashFractionDoubleSign: res.Params.SlashFractionDoubleSign.String(),
		SlashFractionDowntime:   res.Params.SlashFractionDowntime.String(),
	}
	return po
}
