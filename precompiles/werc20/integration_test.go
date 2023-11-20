package werc20_test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v15/precompiles/testutil"
	"github.com/evmos/evmos/v15/precompiles/werc20"
	"github.com/evmos/evmos/v15/precompiles/werc20/testdata"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v15/testutil/integration/evmos/keyring"
	erc20types "github.com/evmos/evmos/v15/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"
)

var _ = Describe("WEVMOS Extension -", func() {
	var (
		WEVMOSContractAddr common.Address
		err                error
		sender             keyring.Key

		// contractData is a helper struct to hold the addresses and ABIs for the
		// different contract instances that are subject to testing here.
		contractData ContractData

		execRevertedCheck testutil.LogCheckArgs
		failCheck         testutil.LogCheckArgs
		passCheck         testutil.LogCheckArgs
	)

	BeforeEach(func() {
		s.SetupTest()

		sender = s.keyring.GetKey(0)

		WEVMOSContractAddr, err = s.factory.DeployContract(
			sender.Priv,
			evmtypes.EvmTxArgs{}, // NOTE: passing empty struct to use default values
			factory.ContractDeploymentData{
				Contract:        testdata.WEVMOSContract,
				ConstructorArgs: []interface{}{},
			},
		)
		Expect(err).ToNot(HaveOccurred(), "failed to deploy contract")

		contractData = ContractData{
			ownerPriv:      sender.Priv,
			erc20Addr:      WEVMOSContractAddr,
			erc20ABI:       testdata.WEVMOSContract.ABI,
			precompileAddr: s.precompile.Address(),
			precompileABI:  s.precompile.ABI,
		}

		// Create the token pair for WEVMOS <-> EVMOS.
		tokenPair := erc20types.NewTokenPair(WEVMOSContractAddr, s.bondDenom, erc20types.OWNER_MODULE)

		precompile, err := werc20.NewPrecompile(
			tokenPair,
			s.network.App.BankKeeper,
			s.network.App.AuthzKeeper,
			s.network.App.TransferKeeper,
		)
		Expect(err).ToNot(HaveOccurred(), "failed to create wevmos extension")
		s.precompile = precompile

		err = s.network.App.EvmKeeper.AddEVMExtensions(s.network.GetContext(), precompile)
		Expect(err).ToNot(HaveOccurred(), "failed to add wevmos extension")

		failCheck = testutil.LogCheckArgs{ABIEvents: s.precompile.Events}
		execRevertedCheck = failCheck.WithErrContains("execution reverted")
		passCheck = failCheck.WithExpPass(true)

		fmt.Println(execRevertedCheck, passCheck)
		err = s.network.NextBlock()
		Expect(err).ToNot(HaveOccurred(), "failed to advance block")

	})

	Context("compatibility with ERC20 extension - ", func() {
		It("should route WEVMOS transfers to the ERC20 extension", func() {
			sender := s.keyring.GetKey(0)
			balance := s.network.App.BankKeeper.GetBalance(s.network.GetContext(), sender.AccAddr, s.bondDenom)
			fmt.Println(balance)

			txArgs, depositArgs := s.getTxAndCallArgs(erc20Call, contractData, werc20.DepositMethod, sender, balance)
			_, ethRes, err := s.factory.CallContractAndCheckLogs(sender.Priv, txArgs, depositArgs, passCheck)
			Expect(err).ToNot(HaveOccurred(), "unexpected result calling contract")

		})
	})
})
