package stride

import (
	"embed"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"log"
)

// Embed memo json file to the executable binary. Needed when importing as dependency.
//
//go:embed memo.json
var memoF embed.FS

const (
	// StrideChannelID is the channel ID for the Stride channel on mainnet.
	StrideChannelID = "channel-25"
	// LiquidStakeEvmosMethod is the method name of the LiquidStakeEvmos method
	LiquidStakeEvmosMethod = "liquidStakeEvmos"
	// OsmoERC20Address is the ERC20 hex address of the Osmosis token on mainnet.
	OsmoERC20Address = "0xFA3C22C069B9556A4B2f7EcE1Ee3B467909f4864"
)

var (
	// WEVMOSAddress is the ERC20 hex address of the WEVMOS token on mainnet.
	WEVMOSAddress = common.HexToAddress("0xD4949664cD82660AaE99bEdc034a0deA8A0bd517")
)

// LiquidStake is a transaction that liquid stakes Evmos using
// a ICS20 transfer with a custom memo field that will trigger Stride's Autopilot middleware
func (p Precompile) LiquidStake(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	erc20Addr, amount, receiverAddress, err := CreateLiquidStakeEvmosPacket(args)
	if err != nil {
		return nil, err
	}

	// TODO: temporary check if the erc20Addr is WEVMOS or the Osmosis token pair
	var coin sdk.Coin
	tokenPairID := p.erc20Keeper.GetTokenPairID(ctx, erc20Addr.String())
	tokenPair, _ := p.erc20Keeper.GetTokenPair(ctx, tokenPairID)
	switch {
	case erc20Addr == WEVMOSAddress:
		coin = sdk.NewCoin("aevmos", sdk.NewIntFromBigInt(amount))
	case tokenPair.Erc20Address == OsmoERC20Address:
		coin = sdk.NewCoin(tokenPair.Denom, sdk.NewIntFromBigInt(amount))
	default:
		return nil, fmt.Errorf("unsupported ERC20 token")
	}

	// Create the memo for the ICS20 transfer
	memo := p.createLiquidStakeMemo(receiverAddress)

	// Build the MsgTransfer with the memo and coin
	msg, err := NewMsgTransfer(StrideChannelID, sdk.AccAddress(origin.Bytes()).String(), receiverAddress, memo, coin)
	if err != nil {
		return nil, err
	}

	// Execute the ICS20 Transfer
	_, err = p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// Emit the IBC transfer Event
	// TODO: Figure out if we want a more custom event here to signal Autopilot usage
	if err = p.EmitIBCTransferEvent(
		ctx,
		stateDB,
		origin,
		msg.Receiver,
		msg.SourcePort,
		msg.SourceChannel,
		msg.Token,
		msg.Memo,
	); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// createLiquidStakeMemo creates the memo for the LiquidStakeEvmos packet
func (p Precompile) createLiquidStakeMemo(receiverAddress string) string {
	// Read the JSON memo from the file
	data, err := memoF.ReadFile("memo.json")
	if err != nil {
		log.Fatalf("Failed to read JSON memo: %v", err)
	}

	// Replace the placeholder with the receiver address
	return fmt.Sprintf(string(data), receiverAddress)
}
