package evm_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

<<<<<<< HEAD
	"github.com/evmos/evmos/v18/app/ante/testutils"
=======
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/client"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v19/app"
	ante "github.com/evmos/evmos/v19/app/ante"
	"github.com/evmos/evmos/v19/encoding"
	"github.com/evmos/evmos/v19/ethereum/eip712"
	"github.com/evmos/evmos/v19/utils"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v19/x/feemarket/types"
>>>>>>> main
)

type AnteTestSuite struct {
	*testutils.AnteTestSuite
	useLegacyEIP712TypedData bool
}

func TestAnteTestSuite(t *testing.T) {
	baseSuite := new(testutils.AnteTestSuite)
	baseSuite.WithLondonHardForkEnabled(true)

	suite.Run(t, &AnteTestSuite{
		AnteTestSuite: baseSuite,
	})

	// Re-run the tests with EIP-712 Legacy encodings to ensure backwards compatibility.
	// LegacyEIP712Extension should not be run with current TypedData encodings, since they are not compatible.
	suite.Run(t, &AnteTestSuite{
		AnteTestSuite:            baseSuite,
		useLegacyEIP712TypedData: true,
	})
}
