package v3

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/evmos/evmos/v10/x/ibc/transfer/types"
)

// at the time of this migration, on mainnet, channels 0 to 36 were open
// so this migration covers those channels only
const openChannels = 36

// MigrateEscrowAccounts updates the IBC transfer escrow accounts type to ModuleAccount
func MigrateEscrowAccounts(ctx sdk.Context, ak types.AccountKeeper) error {
	for i := 0; i <= openChannels; i++ {
		ch := "channel-" + strconv.Itoa(i)
		address := ibctypes.GetEscrowAddress(ibctypes.PortID, ch)

		accountName := fmt.Sprintf("%s/%s", ibctypes.PortID, ch)
		baseAcc := authtypes.NewBaseAccountWithAddress(address)

		// no special permissions defined for the module account
		acc := authtypes.NewModuleAccount(baseAcc, accountName)
		ak.SetModuleAccount(ctx, acc)
	}
	return nil
}
