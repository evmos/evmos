package blocksdk

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
)

func getEthereumSignerFromSDKMsg(msg sdk.Msg) (sdk.AccAddress, uint64, error) {
	// cast the msg to an ethereum msg
	msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
	if !ok {
		return nil, 0, fmt.Errorf("msg of type %T is not an ethereum tx", msg)
	}

	// get the signer from the msg
	signers := msgEthTx.GetSigners()
	if len(signers) != 1 {
		return nil, 0, fmt.Errorf("expected 1 signer, got %d", len(signers))
	}

	// get the nonce from the message

	return signers[0], msgEthTx.AsTransaction().Nonce(), nil
}

func checkEthereumExtensionOptions(tx sdk.Tx) bool {
	// cast the tx to an extension options tx
	txWithExtensions, ok := tx.(authante.HasExtensionOptionsTx)
	if !ok {
		return false
	}

	// get the extension options
	opts := txWithExtensions.GetExtensionOptions()
	if len(opts) == 0 {
		return false
	}

	// check the type url of the first extension option
	typeURL := opts[0].GetTypeUrl()
	return typeURL == ethereumSignatureExtensionOption
}
