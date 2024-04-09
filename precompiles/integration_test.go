package precompiles_test

import (
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	ics20precompile "github.com/evmos/evmos/v16/precompiles/ics20"
	stakingprecompile "github.com/evmos/evmos/v16/precompiles/staking"
	"github.com/evmos/evmos/v16/precompiles/testutil"
	testfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"

	//nolint:revive // it's common practice to use the global imports for Ginkgo and Gomega
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // it's common practice to use the global imports for Ginkgo and Gomega
	. "github.com/onsi/gomega"
)

func TestApprovalBehavior(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Approval Behavior Suite")
}

var _ = Describe("Checking approval behavior -", func() {
	var (
		keyring testkeyring.Keyring
		network *testnetwork.UnitTestNetwork // using the unit test network here in order to enable keeper access to instantiate the precompiles
		handler grpc.Handler
		factory testfactory.TxFactory

		stakingPrecompileAddress common.Address
		stakingABI               abi.ABI

		ics20PrecompileAddress common.Address
		ics20ABI               abi.ABI

		erc20PrecompileAddress common.Address
		erc20ABI               abi.ABI
	)

	BeforeEach(func() {
		keyring = testkeyring.New(2)
		network = testnetwork.NewUnitTestNetwork(
			testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		)
		handler = grpc.NewIntegrationHandler(network)
		factory = testfactory.New(network, handler)

		stakingPrecompileAddress = common.HexToAddress(stakingprecompile.PrecompileAddress)
		ics20PrecompileAddress = ics20precompile.Precompile{}.Address()

		var err error
		stakingABI, err = stakingprecompile.LoadABI()
		Expect(err).ToNot(HaveOccurred(), "failed to load staking precompile ABI")

		ics20ABI, err = ics20precompile.LoadABI()
		Expect(err).ToNot(HaveOccurred(), "failed to load ics20 precompile ABI")

		erc20PrecompileAddress = erc20precompile.Precompile{}.Address()
		erc20ABI = erc20precompile.GetABI()
		Expect(err).ToNot(HaveOccurred(), "failed to load erc20 precompile ABI")
	})

	Context("approving", func() {
		var (
			granter, grantee testkeyring.Key
		)

		BeforeEach(func() {
			granter = keyring.GetKey(0)
			grantee = keyring.GetKey(1)

		})

		Context("with no prior authorization", func() {
			BeforeEach(func() {
				authzs, err := handler.GetAuthorizations(
					granter.AccAddr.String(),
					grantee.AccAddr.String(),
				)
				Expect(err).ToNot(HaveOccurred(), "failed to get authorizations")
				Expect(authzs).To(HaveLen(0), "expected no previous authorizations to exist")
			})

			It("should create an unlimited approval with the max uint 256 amount (staking)", func() {
				_, _, err := factory.CallContractAndCheckLogs(
					granter.Priv,
					evmtypes.EvmTxArgs{
						To: &stakingPrecompileAddress,
					},
					testfactory.CallArgs{
						ContractABI: stakingABI,
						MethodName:  "approve",
						Args:        []interface{}{grantee.Addr, abi.MaxUint256, []string{stakingprecompile.DelegateMsg}},
					},
					testutil.LogCheckArgs{
						ABIEvents: stakingABI.Events,
						ExpEvents: []string{"Approval"},
						ExpPass:   true,
					},
				)
				Expect(err).ToNot(HaveOccurred(), "expected different result calling contract")

				authzs, err := handler.GetAuthorizations(
					grantee.AccAddr.String(),
					granter.AccAddr.String(),
				)
				Expect(err).ToNot(HaveOccurred(), "failed to get authorizations")
				Expect(authzs).To(HaveLen(1), "expected authorization to be created")

				stakeAuthz, ok := authzs[0].(*stakingtypes.StakeAuthorization)
				Expect(ok).To(BeTrue(), "expected authorization to be a stake authorization")
				Expect(stakeAuthz.GetMaxTokens()).To(
					BeNil(),
					"expected nil amount as a sign of an unlimited stake authorization",
				)
			})

			It("should create an unlimited approval with the max uint 256 amount (ics20)", func() {
				_, _, err := factory.CallContractAndCheckLogs(
					granter.Priv,
					evmtypes.EvmTxArgs{
						To: &ics20PrecompileAddress,
					},
					testfactory.CallArgs{
						ContractABI: ics20ABI,
						MethodName:  "approve",
						Args:        []interface{}{grantee.Addr, abi.MaxUint256},
					},
					testutil.LogCheckArgs{
						ABIEvents: ics20ABI.Events,
						ExpEvents: []string{"Approval"},
						ExpPass:   true,
					},
				)
				Expect(err).ToNot(HaveOccurred(), "expected different result calling contract")

				authzs, err := handler.GetAuthorizations(
					grantee.AccAddr.String(),
					granter.AccAddr.String(),
				)
				Expect(err).ToNot(HaveOccurred(), "failed to get authorizations")
				Expect(authzs).To(HaveLen(1), "expected authorization to be created")

				ics20Authz, ok := authzs[0].(*transfertypes.TransferAuthorization)
				Expect(ok).To(BeTrue(), "expected authorization to be an ics20 authorization")
				Expect(ics20Authz.Allocations).To(
					HaveLen(1),
					"expected one allocation in ics20 authorization",
				)
				Expect(ics20Authz.Allocations[0].SpendLimit.String()).To(
					Equal(""),
					"expected no amount as a sign of an unlimited ics20 authorization",
				)
			})
		})
	})
})
