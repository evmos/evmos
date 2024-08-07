package cosmos_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

<<<<<<< HEAD
	"github.com/evmos/evmos/v19/app/ante/testutils"
	storetypes "cosmossdk.io/store/types"
=======
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/evmos/evmos/v19/app"
	"github.com/evmos/evmos/v19/app/ante"
	evmante "github.com/evmos/evmos/v19/app/ante/evm"
	"github.com/evmos/evmos/v19/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v19/encoding"
	"github.com/evmos/evmos/v19/ethereum/eip712"
	"github.com/evmos/evmos/v19/testutil"
	"github.com/evmos/evmos/v19/types"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v19/x/feemarket/types"
>>>>>>> main
)

type AnteTestSuite struct {
	*testutils.AnteTestSuite
}

func TestAnteTestSuite(t *testing.T) {
	baseSuite := new(testutils.AnteTestSuite)
	baseSuite.WithLondonHardForkEnabled(true)
	baseSuite.WithFeemarketEnabled(true)

	suite.Run(t, &AnteTestSuite{baseSuite})
}
