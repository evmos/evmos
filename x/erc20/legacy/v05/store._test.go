package v05_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v2/testing/simapp"
)

func TestTokenPairKeysMigration(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	erc20Key := sdk.NewKVStoreKey("erc20")
	ctx := testutil.DefaultContext(erc20Key, sdk.)

}
