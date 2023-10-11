package stride

import (
	"embed"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/authorization"
	"github.com/evmos/evmos/v14/precompiles/ics20"
	"log"
	"time"
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
	contract *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	sender, erc20Addr, amount, receiverAddress, err := CreateLiquidStakeEvmosPacket(args)
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

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided sender address is different from the origin address,
	// return an error because is a forbidden operation
	if contract.CallerAddress == sender {
		sender = origin
	} else if origin != sender {
		return nil, fmt.Errorf(ics20.ErrDifferentOriginFromSender, origin.String(), sender.String())
	}

	// Create the memo for the ICS20 transfer
	memo := p.createLiquidStakeMemo(receiverAddress)

	// Build the MsgTransfer with the memo and coin
	msg, err := NewMsgTransfer(StrideChannelID, sdk.AccAddress(sender.Bytes()).String(), receiverAddress, memo, coin)
	if err != nil {
		return nil, err
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	// and the sender is the origin
	var (
		expiration *time.Time
		auth       authz.Authorization
		resp       *authz.AcceptResponse
	)
	if contract.CallerAddress != origin {
		// check if authorization exists
		auth, expiration, err = authorization.CheckAuthzExists(ctx, p.AuthzKeeper, contract.CallerAddress, origin, ics20.TransferMsgURL)
		if err != nil {
			return nil, fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, contract.CallerAddress, origin)
		}

		// Accept the grant and return an error if the grant is not accepted
		resp, err = ics20.AcceptGrant(ctx, contract.CallerAddress, origin, msg, auth)
		if err != nil {
			return nil, err
		}
	}

	// Execute the ICS20 Transfer
	_, err = p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// Update grant only if is needed
	if contract.CallerAddress != origin {
		// accepts and updates the grant adjusting the spending limit
		if err = ics20.UpdateGrant(ctx, p.AuthzKeeper, contract.CallerAddress, origin, expiration, resp); err != nil {
			return nil, err
		}
	}

	// Emit the IBC transfer Event
	if err = ics20.EmitIBCTransferEvent(ctx, stateDB, p.ABI.Events, sender, p.Address(), msg); err != nil {
		return nil, err
	}

	// Emit the custom LiquidStake Event
	if err = p.EmitLiquidStakeEvent(ctx, stateDB, sender, erc20Addr, amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// createLiquidStakeMemo creates the memo for the LiquidStake packet
func (p Precompile) createLiquidStakeMemo(receiverAddress string) string {
	// Read the JSON memo from the file
	data, err := memoF.ReadFile("memo.json")
	if err != nil {
		log.Fatalf("Failed to read JSON memo: %v", err)
	}

	// Replace the placeholder with the receiver address
	return fmt.Sprintf(string(data), receiverAddress)
}
