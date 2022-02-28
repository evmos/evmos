package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	channelkeeper "github.com/cosmos/ibc-go/v3/modules/core/04-channel/keeper"
	ibcante "github.com/cosmos/ibc-go/v3/modules/core/ante"

	ethante "github.com/tharsis/ethermint/app/ante"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	vestingtypes "github.com/tharsis/evmos/x/vesting/types"
)

// HandlerOptions defines the list of module keepers required to run the Evmos
// AnteHandler decorators.
type HandlerOptions struct {
	AccountKeeper    evmtypes.AccountKeeper
	BankKeeper       evmtypes.BankKeeper
	IBCChannelKeeper channelkeeper.Keeper
	FeeMarketKeeper  evmtypes.FeeMarketKeeper
	StakingKeeper    vestingtypes.StakingKeeper
	EvmKeeper        ethante.EVMKeeper
	FeegrantKeeper   ante.FeegrantKeeper
	SignModeHandler  authsigning.SignModeHandler
	SigGasConsumer   func(meter sdk.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
}

// Validate checks if the keepers are defined
func (options HandlerOptions) Validate() error {
	if options.AccountKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.StakingKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "staking keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.FeeMarketKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "fee market keeper is required for AnteHandler")
	}
	if options.EvmKeeper == nil {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "evm keeper is required for AnteHandler")
	}
	return nil
}

// newCosmosAnteHandler creates the default ante handler for Ethereum transactions
func newEthAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		ethante.NewEthSetUpContextDecorator(options.EvmKeeper), // outermost AnteDecorator. SetUpContext must be called first
		ethante.NewEthMempoolFeeDecorator(options.EvmKeeper),   // Check eth effective gas price against minimal-gas-prices
		ethante.NewEthValidateBasicDecorator(options.EvmKeeper),
		ethante.NewEthSigVerificationDecorator(options.EvmKeeper),
		ethante.NewEthAccountVerificationDecorator(options.AccountKeeper, options.BankKeeper, options.EvmKeeper),
		ethante.NewEthGasConsumeDecorator(options.EvmKeeper),
		ethante.NewCanTransferDecorator(options.EvmKeeper),
		NewEthVestingTransactionDecorator(options.AccountKeeper),
		ethante.NewEthIncrementSenderSequenceDecorator(options.AccountKeeper), // innermost AnteDecorator.
	)
}

// newCosmosAnteHandler creates the default ante handler for Cosmos transactions
func newCosmosAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		ethante.RejectMessagesDecorator{}, // reject MsgEthereumTxs
		ante.NewSetUpContextDecorator(),
		ante.NewRejectExtensionOptionsDecorator(),
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		NewVestingDelegationDecorator(options.AccountKeeper, options.StakingKeeper),
		NewValidatorCommissionDecorator(),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewAnteDecorator(options.IBCChannelKeeper),
	)
}

// newCosmosAnteHandlerEip712 creates the ante handler for transactions signed with EIP712
func newCosmosAnteHandlerEip712(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		ethante.RejectMessagesDecorator{}, // reject MsgEthereumTxs
		ante.NewSetUpContextDecorator(),
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		NewVestingDelegationDecorator(options.AccountKeeper, options.StakingKeeper),
		NewValidatorCommissionDecorator(),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		// Note: signature verification uses EIP instead of the cosmos signature validator
		ethante.NewEip712SigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewAnteDecorator(options.IBCChannelKeeper),
	)
}
