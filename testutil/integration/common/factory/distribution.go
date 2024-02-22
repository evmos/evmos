package factory

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (tf *IntegrationTxFactory) SetWithdrawAddress(delegatorPriv cryptotypes.PrivKey, withdrawerAddr sdk.AccAddress) error {
	delegatorAccAddr := sdk.AccAddress(delegatorPriv.PubKey().Address())

	msgDelegate := distrtypes.NewMsgSetWithdrawAddress(
		delegatorAccAddr,
		withdrawerAddr,
	)

	resp, err := tf.ExecuteCosmosTx(delegatorPriv, CosmosTxArgs{
		Msgs: []sdk.Msg{msgDelegate},
	})

	if resp.Code != 0 {
		err = fmt.Errorf("received error code %d on SetWithdrawAddress transaction. Logs: %s", resp.Code, resp.Log)
	}

	return err
}
