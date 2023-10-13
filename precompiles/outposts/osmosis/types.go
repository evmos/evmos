package osmosis

import (
	"fmt"
	"math/big"
	"strings"

	errorsmod "cosmossdk.io/errors"

	"github.com/evmos/evmos/v14/precompiles/authorization"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

// TODO: This is the function we will use for V1 of the Osmosis swap function.
// CreateSwapPacketDataV1 creates the packet data for the Osmosis swap function.
//
//	func CreateSwapPacketDataV1(args []interface{}, ctx sdk.Context, bankKeeper erc20types.BankKeeper, erc20Keeper erckeeper.Keeper) (*big.Int, string, string, string, error) {
//		if len(args) != 4 {
//			return nil, "", "", "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
//		}
//
//		amount, ok := args[0].(*big.Int)Yeah
//		if !ok {
//			return nil, "", "", "", fmt.Errorf("invalid amount: %v", args[0])
//		}
//
//		receiverAddress, ok := args[1].(string)
//		if !ok {
//			return nil, "", "", "", fmt.Errorf("invalid receiver address: %v", args[1])
//		}
//
//		inputContract, ok := args[2].(common.Address)
//		if !ok {
//			return nil, "", "", "", fmt.Errorf("invalid input denom: %v", args[2])
//		}
//
//		inputVoucher, found := erc20Keeper.GetTokenPair(ctx, erc20Keeper.GetERC20Map(ctx, inputContract))
//		if !found {
//			return nil, "", "", "", fmt.Errorf("invalid input denom: %v", inputContract.String())
//		}
//
//		inputDenomMetadata, found := bankKeeper.GetDenomMetaData(ctx, inputVoucher.Denom)
//		if !found {
//			return nil, "", "", "", fmt.Errorf("invalid input denom: %v", inputContract.String())
//		}
//
//		fmt.Println(inputDenomMetadata)
//
//		outputContract, ok := args[3].(common.Address)
//		if !ok {
//			return nil, "", "", "", fmt.Errorf("invalid output denom: %v", args[3])
//		}
//
//		outputDenomMetadata, found := bankKeeper.GetDenomMetaData(ctx, outputContract.String())
//		if !found {
//			return nil, "", "", "", fmt.Errorf("invalid input denom: %v", inputContract.String())
//		}
//
//		// TODO: is this the right way to extract the prefix
//		prefix, _, err := bech32.DecodeAndConvert(receiverAddress)
//		if err != nil {
//			return nil, "", "", "", fmt.Errorf("invalid receiver address: %v", err)
//		}
//
//		fmt.Println(prefix)
//
//		return amount, inputDenomMetadata.Base, outputDenomMetadata.Base, receiverAddress, nil
//	}

type Twap struct {
	SlippagePercentage string `json:"slippage_percentage"`
	WindowSeconds      string `json:"window_seconds"`
}

type Slippage struct {
	Twap Twap `json:"twap"`
}

type OsmosisSwap struct {
	OutputDenom      string   `json:"output_denom"`
	Slippage         Slippage `json:"slippage"`
	receiver         string
	onFailedDelivery string `json:"on_failed_delivery"`
	nextMemo         string `json:"next_memo"`
}

type Msg struct {
	OsmosisSwap OsmosisSwap `json:"osmosis_swap",omitempty`
}

type Memo struct {
	Contract string `json:"contract"`
	Msg      Msg    `json:"msg"`
}

type RawPacketMetadata struct {
	Memo Memo `json:"memo"`
}

// ParseSwapPacketData parses the packet data for the Osmosis swap function.
func ParseSwapPacketData(args []interface{}) (sender common.Address, input common.Address, output common.Address, amount *big.Int, receiver string, err error) {
	if len(args) != 5 {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 5, len(args))
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf("invalid sender address: %v", args[0])
	}

	input, ok = args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf("invalid input denom: %v", args[1])
	}

	output, ok = args[2].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf("invalid output denom: %v", args[2])
	}

	amount, ok = args[3].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf("invalid amount: %v", args[3])
	}

	receiver, ok = args[4].(string)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf("invalid receiver address: %v", args[4])
	}

	return sender, input, output, amount, receiver, nil
}

// NewMsgTransfer creates a new MsgTransfer
func NewMsgTransfer(sourcePort, sourceChannel, senderAddress, receiverAddress, memo string, coin sdk.Coin) (*transfertypes.MsgTransfer, error) {
	// TODO: what are some sensible defaults here
	timeoutHeight := clienttypes.NewHeight(100, 100)
	msg := transfertypes.NewMsgTransfer(
		sourcePort,
		sourceChannel,
		coin,
		senderAddress,
		receiverAddress,
		timeoutHeight,
		0,
		memo,
	)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}

// NewTransferAuthorization returns a new transfer authorization authz type from the given arguments.
// Pre-populates the channel and port id to only work with Osmosis.
func NewTransferAuthorization(args []interface{}) (common.Address, *transfertypes.TransferAuthorization, error) {
	if len(args) != 3 {
		return common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	grantee, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, nil, fmt.Errorf(authorization.ErrInvalidGrantee, args[0])
	}

	spendLimit, ok := args[1].([]cmn.Coin)
	if !ok {
		return common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidType, "spendLimit", cmn.Coin{}, args[1])
	}

	allowList, ok := args[2].([]string)
	if !ok {
		return common.Address{}, nil, fmt.Errorf(cmn.ErrInvalidType, "allowList", []string{}, args[2])
	}

	spendLimitCoins := make(sdk.Coins, len(spendLimit))
	for is, sl := range spendLimit {
		spendLimitCoins[is] = sdk.Coin{
			Amount: math.NewIntFromBigInt(sl.Amount),
			Denom:  sl.Denom,
		}
	}

	allocations := make([]transfertypes.Allocation, 1)
	allocations[0] = transfertypes.Allocation{
		SourcePort:    transfertypes.PortID,
		SourceChannel: OsmosisChannelID,
		SpendLimit:    spendLimitCoins,
		AllowList:     allowList,
	}

	transferAuthz := &transfertypes.TransferAuthorization{Allocations: allocations}
	if err := transferAuthz.ValidateBasic(); err != nil {
		return common.Address{}, nil, err
	}

	return grantee, transferAuthz, nil
}

// AccAddressFromBech32 creates an AccAddress from a Bech32 string.
func AccAddressFromBech32(address string, bech32prefix string) (addr sdk.AccAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return sdk.AccAddress{}, fmt.Errorf("empty address string is not allowed")
	}

	bz, err := sdk.GetFromBech32(address, bech32prefix)
	if err != nil {
		return nil, err
	}

	err = sdk.VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return sdk.AccAddress(bz), nil
}

// checkAllowanceArgs checks the arguments for the Increase / Decrease Allowance function.
func checkAllowanceArgs(args []interface{}) (common.Address, string, *big.Int, error) {
	if len(args) != 3 {
		return common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	grantee, ok := args[0].(common.Address)
	if !ok || grantee == (common.Address{}) {
		return common.Address{}, "", nil, fmt.Errorf(authorization.ErrInvalidGrantee, args[0])
	}

	denom, ok := args[1].(string)
	if !ok {
		return common.Address{}, "", nil, errorsmod.Wrapf(transfertypes.ErrInvalidDenomForTransfer, cmn.ErrInvalidDenom, args[1])
	}

	amount, ok := args[4].(*big.Int)
	if !ok || amount == nil {
		return common.Address{}, "", nil, errorsmod.Wrapf(transfertypes.ErrInvalidAmount, cmn.ErrInvalidAmount, args[2])
	}

	return grantee, denom, amount, nil
}
