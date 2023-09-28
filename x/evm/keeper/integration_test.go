package keeper_test

import (
	"encoding/json"
	"math/big"

	"cosmossdk.io/math"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/contracts"
	"github.com/evmos/evmos/v14/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v14/encoding"
	"github.com/evmos/evmos/v14/testutil"

	"github.com/evmos/evmos/v14/testutil/integration/factory"
	"github.com/evmos/evmos/v14/testutil/integration/grpc"
	testkeyring "github.com/evmos/evmos/v14/testutil/integration/keyring"
	"github.com/evmos/evmos/v14/testutil/integration/network"

	utiltx "github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/utils"
	evmtypes "github.com/evmos/evmos/v14/x/evm/types"
	"github.com/evmos/evmos/v14/x/feemarket/types"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type IntegrationTestSuite struct {
	network     network.Network
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring
}

var _ = Describe("Handling a MsgEthereumTx message", Label("EVM"), Ordered, func() {
	var s *IntegrationTestSuite

	BeforeAll(func() {
		keyring := testkeyring.New(3)
		integrationNetwork := network.New(
			network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		)
		grpcHandler := grpc.NewIntegrationHandler(integrationNetwork)
		txFactory := factory.New(integrationNetwork, grpcHandler)
		s = &IntegrationTestSuite{
			network:     integrationNetwork,
			factory:     txFactory,
			grpcHandler: grpcHandler,
			keyring:     keyring,
		}
	})

	When("the params have default values", Ordered, func() {
		BeforeAll(func() {
			// Set params to default values
			defaultParams := evmtypes.DefaultParams()
			err := s.network.UpdateEvmParams(defaultParams)
			Expect(err).To(BeNil())
		})

		It("performs a transfer call", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			receiver := s.keyring.GetKey(1)
			txArgs := evmtypes.EvmTxArgs{
				To:     &receiver.Addr,
				Amount: big.NewInt(1000),
			}

			res, err := s.factory.ExecuteEthTx(senderPriv, txArgs)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
		})

		It("performs a contract deployment and contract call", func() {
			senderPriv := s.keyring.GetPrivKey(0)
			constructorArgs := []interface{}{"coin", "token", uint8(18)}
			compiledContract := contracts.ERC20MinterBurnerDecimalsContract
			contractAddr, err := s.factory.DeployContract(
				senderPriv,
				compiledContract,
				constructorArgs...,
			)
			Expect(err).To(BeNil())
			Expect(contractAddr).ToNot(Equal(common.Address{}))

			txArgs := evmtypes.EvmTxArgs{
				To: &contractAddr,
			}
			callArgs := factory.CallArgs{
				ContractABI: compiledContract.ABI,
				MethodName:  "mint",
				Args:        []interface{}{s.keyring.GetAddr(1), big.NewInt(1e18)},
			}
			res, err := s.factory.ExecuteContractCall(senderPriv, txArgs, callArgs)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
		})
	})
})
