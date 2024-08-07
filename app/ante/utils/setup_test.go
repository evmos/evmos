package utils_test

import (
	"testing"

	"github.com/evmos/evmos/v19/app/ante/testutils"
	"github.com/stretchr/testify/suite"
<<<<<<< HEAD
=======

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v19/app"
	"github.com/evmos/evmos/v19/app/ante"
	"github.com/evmos/evmos/v19/encoding"
	"github.com/evmos/evmos/v19/ethereum/eip712"
	"github.com/evmos/evmos/v19/testutil"
	"github.com/evmos/evmos/v19/utils"
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

	suite.Run(t, &AnteTestSuite{
		AnteTestSuite: baseSuite,
	})
}
