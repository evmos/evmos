package evm

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/x/evm/keeper"
	"github.com/evmos/evmos/v15/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

// EthAccountVerificationDecorator validates an account balance checks
type EthAccountVerificationDecorator struct {
	ak        evmtypes.AccountKeeper
	evmKeeper EVMKeeper
}

// NewEthAccountVerificationDecorator creates a new EthAccountVerificationDecorator
func NewEthAccountVerificationDecorator(ak evmtypes.AccountKeeper, ek EVMKeeper) EthAccountVerificationDecorator {
	return EthAccountVerificationDecorator{
		ak:        ak,
		evmKeeper: ek,
	}
}

// AnteHandle validates checks that the sender balance is greater than the total transaction cost.
// The account will be set to store if it doesn't exist, i.e. cannot be found on store.
// This AnteHandler decorator will fail if:
// - any of the msgs is not a MsgEthereumTx
// - from address is empty
// - account balance is lower than the transaction cost
func (avd EthAccountVerificationDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	for _, msg := range tx.GetMsgs() {
		_, txData, from, err := evmtypes.UnpackEthMsg(msg)
		if err != nil {
			return ctx, err
		}

		if err := VerifyAccountBalance(ctx, avd.ak, avd.evmKeeper, from, txData); err != nil {
			return ctx, err
		}
	}
	return next(ctx, tx, simulate)
}

func VerifyAccountBalance(
	ctx sdk.Context,
	accountKeeper evmtypes.AccountKeeper,
	evmKeeper EVMKeeper,
	from sdk.AccAddress,
	txData evmtypes.TxData,
) error {
	// check whether the sender address is EOA
	fromAddr := common.BytesToAddress(from)
	acct := evmKeeper.GetAccount(ctx, fromAddr)

	if acct == nil {
		acc := accountKeeper.NewAccountWithAddress(ctx, from)
		accountKeeper.SetAccount(ctx, acc)
		acct = statedb.NewEmptyAccount()
	} else if acct.IsContract() {
		return errorsmod.Wrapf(errortypes.ErrInvalidType,
			"the sender is not EOA: address %s, codeHash <%s>", fromAddr, acct.CodeHash)
	}

	if err := keeper.CheckSenderBalance(sdkmath.NewIntFromBigInt(acct.Balance), txData); err != nil {
		return errorsmod.Wrap(err, "failed to check sender balance")
	}

	return nil
}
