package v17_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	v17 "github.com/evmos/evmos/v16/app/upgrades/v17"
	testfactory "github.com/evmos/evmos/v16/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	testutils "github.com/evmos/evmos/v16/testutil/integration/evmos/utils"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"

	//nolint:revive // dot-imports are okay for Ginkgo BDD
	. "github.com/onsi/ginkgo/v2"
	//nolint:revive // dot-imports are okay for Gomega assertions
	. "github.com/onsi/gomega"
)

// TestSTRv2Migration runs the Ginkgo BDD tests for the migration logic
// associated with the Single Token Representation v2.
func TestSTRv2Migration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "STR v2 Migration Suite")
}

type ConvertERC20CoinsTestSuite struct {
	keyring testkeyring.Keyring
	network *network.UnitTestNetwork
	handler grpc.Handler
	factory testfactory.TxFactory

	// erc20Contract is the address of the deployed ERC-20 contract for testing purposes.
	erc20Contract common.Address
	// nativeTokenPair is a registered token pair for a native Coin.
	nativeTokenPair erc20types.TokenPair
	// nonNativeTokenPair is a registered token pair for an ERC-20 native asset.
	nonNativeTokenPair erc20types.TokenPair
	// wevmosContract is the address of the deployed WEVMOS contract for testing purposes.
	wevmosContract common.Address
}

// NOTE: For these tests it's mandatory to run them ORDERED!
var _ = When("testing the STR v2 migration", Ordered, func() {
	var (
		ts *ConvertERC20CoinsTestSuite

		// unconverted is the amount of native coins that have not been converted to ERC-20.
		unconverted = int64(500)
		// converted is the amount of native coins that have been converted to ERC-20.
		converted = int64(100)
	)

	BeforeAll(func() {
		// NOTE: In the setup function we are creating a custom genesis state for the integration network
		// which contains balances for accounts in different denominations, both in native as well as ERC-20 representation.
		var err error
		ts, err = NewConvertERC20CoinsTestSuite()
		Expect(err).ToNot(HaveOccurred(), "failed to create test suite")
	})

	When("checking the genesis state of the network", Ordered, func() {
		It("should include all expected accounts", func() {
			expectedAccounts := ts.keyring.GetAllAccAddrs()
			expectedAccounts = append(expectedAccounts, bech32WithERC20s, SmartContractAddress)

			for _, acc := range expectedAccounts {
				got, err := ts.handler.GetAccount(acc.String())
				Expect(err).ToNot(HaveOccurred(), "failed to get account")
				Expect(got).ToNot(BeNil(), "expected non-nil account")
			}
		})

		It("should have the expected initial balances", func() {
			// We check that the ERC20 converted coins have been added back to the bank balance.
			//
			// NOTE: We are deliberately ONLY checking the balance of the XMPL coin, because the AEVMOS balance was changed
			// through paying transaction fees and they are not affected by the migration.
			err := testutils.CheckBalances(ts.handler, []banktypes.Balance{
				{
					Address: ts.keyring.GetAccAddr(erc20Deployer).String(),
					Coins: sdk.NewCoins(
						sdk.NewCoin(AEVMOS, network.PrefundedAccountInitialBalance),
					),
				},
				{
					Address: bech32WithERC20s.String(),
					Coins: sdk.NewCoins(
						sdk.NewCoin(AEVMOS, network.PrefundedAccountInitialBalance),
						sdk.NewInt64Coin(XMPL, unconverted),
					),
				},
			})
			Expect(err).ToNot(HaveOccurred(), "expected different balances")
		})

		It("should have registered a native token pair", func() {
			res, err := ts.handler.GetTokenPairs()
			Expect(err).ToNot(HaveOccurred(), "failed to get token pairs")
			Expect(res.TokenPairs).To(HaveLen(1), "unexpected number of token pairs")
			Expect(res.TokenPairs[0].Denom).To(Equal(XMPL), "expected different denom")
			Expect(res.TokenPairs[0].IsNativeCoin()).To(BeTrue(), "expected token pair to be for a native coin")

			Expect(res.TokenPairs[0].Erc20Address).To(
				Equal(common.BytesToAddress(SmartContractAddress.Bytes()).String()),
				"expected different ERC-20 contract",
			)

			// Assign the native token pair to the test suite for later use.
			ts.nativeTokenPair = res.TokenPairs[0]
		})

		It("should show separate ERC-20 and bank balances", func() {
			xmplBalance, err := ts.handler.GetBalance(bech32WithERC20s, XMPL)
			Expect(err).ToNot(HaveOccurred(), "failed to get XMPL balance")
			Expect(xmplBalance.Balance.Amount.Int64()).To(
				Equal(unconverted),
				"expected different XMPL balance",
			)

			// Test that the ERC-20 contract for the IBC native coin has the correct user balance after genesis.
			balance, err := GetERC20BalanceForAddr(
				ts.factory,
				ts.keyring.GetPrivKey(erc20Deployer),
				accountWithERC20s,
				ts.nativeTokenPair.GetERC20Contract(),
			)
			Expect(err).ToNot(HaveOccurred(), "failed to query ERC20 balance")
			Expect(balance.String()).To(
				Equal(big.NewInt(converted).String()),
				"expected different ERC-20 balance after genesis",
			)
		})

		It("should have a balance of escrowed tokens in the ERC-20 module account", func() {
			balancesRes, err := ts.handler.GetAllBalances(authtypes.NewModuleAddress(erc20types.ModuleName))
			Expect(err).ToNot(HaveOccurred(), "failed to get balances")
			Expect(balancesRes.Balances).ToNot(BeNil(), "expected non-nil balances")
			Expect(balancesRes.Balances).ToNot(BeEmpty(), "expected non-empty balances")
		})
	})

	When("preparing the network state", Ordered, func() {
		It("should run the preparation without an error", func() {
			var err error
			ts, err = PrepareNetwork(ts)
			Expect(err).ToNot(HaveOccurred(), "failed to prepare network state")
		})

		It("should have registered a non-native token pair", func() {
			res, err := ts.handler.GetTokenPairs()
			Expect(err).ToNot(HaveOccurred(), "failed to get token pairs")
			Expect(res.TokenPairs).To(HaveLen(2), "unexpected number of token pairs")
			Expect(res.TokenPairs).To(ContainElement(ts.nonNativeTokenPair), "non-native token pair not found")
		})

		It("should have minted ERC-20 tokens for the contract deployer", func() {
			balance, err := GetERC20Balance(ts.factory, ts.keyring.GetPrivKey(erc20Deployer), ts.erc20Contract)
			Expect(err).ToNot(HaveOccurred(), "failed to query ERC-20 balance")
			Expect(balance).To(Equal(mintAmount), "expected different balance after minting ERC-20")
		})
	})

	When("running the migration", Ordered, func() {
		// balancePre is the balance of the account having some WEVMOS tokens before the migration.
		//
		// NOTE: we are checking the balances of the account before the migration to compare
		// them with the balances after the migration to check that the WEVMOS tokens
		// have been correctly unwrapped.
		var balancePre *sdk.Coin

		BeforeAll(func() {
			balancePreRes, err := ts.handler.GetBalance(ts.keyring.GetAccAddr(erc20Deployer), AEVMOS)
			Expect(err).ToNot(HaveOccurred(), "failed to check balances")
			balancePre = balancePreRes.Balance
		})

		It("should succeed", func() {
			logger := ts.network.GetContext().Logger().With("upgrade")

			// Convert the coins back using the upgrade util
			err := v17.ConvertToNativeCoinExtensions(
				ts.network.GetContext(),
				logger,
				ts.network.App.AccountKeeper,
				ts.network.App.BankKeeper,
				ts.network.App.Erc20Keeper,
				ts.wevmosContract,
			)
			Expect(err).ToNot(HaveOccurred(), "failed to run migration")

			err = ts.network.NextBlock()
			Expect(err).ToNot(HaveOccurred(), "failed to execute block")
		})

		It("should have converted the ERC-20s back to the native representation", func() {
			// We check that the ERC20 converted coins have been added back to the bank balance.
			err := testutils.CheckBalances(ts.handler, []banktypes.Balance{
				{
					Address: bech32WithERC20s.String(),
					Coins: sdk.NewCoins(
						sdk.NewCoin(AEVMOS, network.PrefundedAccountInitialBalance),
						sdk.NewInt64Coin(XMPL, unconverted+converted),
					),
				},
			})
			Expect(err).ToNot(HaveOccurred(), "expected different balances")
		})

		It("should have converted WEVMOS back to the base denomination", func() {
			// We are checking that the WEVMOS tokens have been converted back to the base denomination.
			balancePostRes, err := ts.handler.GetBalance(ts.keyring.GetAccAddr(erc20Deployer), AEVMOS)
			Expect(err).ToNot(HaveOccurred(), "failed to check balances")
			Expect(balancePostRes.Balance.String()).To(Equal(balancePre.AddAmount(sentWEVMOS).String()), "expected different balance after converting WEVMOS back to unwrapped denom")
		})

		It("should have registered only the native token pair as an active precompile", func() {
			// We check that the token pair was registered as an active precompile.
			evmParamsRes, err := ts.handler.GetEvmParams()
			Expect(err).ToNot(HaveOccurred(), "failed to get EVM params")
			Expect(evmParamsRes.Params.ActivePrecompiles).To(
				ContainElement(ts.nativeTokenPair.GetERC20Contract().String()),
				"expected precompile to be registered",
			)
			Expect(evmParamsRes.Params.ActivePrecompiles).ToNot(
				ContainElement(ts.nonNativeTokenPair.GetERC20Contract().String()),
				"expected no precompile to be registered for non-native token pairs",
			)
		})

		It("should be possible to query the account balance either through the bank or the ERC-20 contract", func() {
			// NOTE: We check that the ERC20 contract for the native token pair can still be called,
			// even though the original contract code was deleted, and it is now re-deployed
			// as a precompiled contract.
			balance, err := GetERC20BalanceForAddr(
				ts.factory,
				ts.keyring.GetPrivKey(erc20Deployer),
				accountWithERC20s,
				ts.nativeTokenPair.GetERC20Contract(),
			)
			Expect(err).ToNot(HaveOccurred(), "failed to query ERC20 balance")
			Expect(balance.Int64()).To(Equal(unconverted+converted), "expected different balance after converting ERC20")

			balanceRes, err := ts.handler.GetBalance(bech32WithERC20s, ts.nativeTokenPair.Denom)
			Expect(err).ToNot(HaveOccurred(), "failed to check balances")
			Expect(balanceRes.Balance.Amount.Int64()).To(Equal(unconverted+converted), "expected different balance after converting ERC20")
		})

		It("should have removed all balances from the ERC-20 module account", func() {
			balancesRes, err := ts.handler.GetAllBalances(authtypes.NewModuleAddress(erc20types.ModuleName))
			Expect(err).ToNot(HaveOccurred(), "failed to get balances")
			Expect(balancesRes.Balances.IsZero()).To(BeTrue(), "expected different balance for module account")
		})

		It("should not have converted the native ERC-20s", func() {
			balance, err := GetERC20Balance(ts.factory, ts.keyring.GetPrivKey(erc20Deployer), ts.nonNativeTokenPair.GetERC20Contract())
			Expect(err).ToNot(HaveOccurred(), "failed to query ERC20 balance")
			Expect(balance).To(Equal(mintAmount), "expected different balance after converting ERC20")
		})

		It("should have withdrawn all WEVMOS tokens", func() {
			balance, err := GetERC20Balance(ts.factory, ts.keyring.GetPrivKey(erc20Deployer), ts.wevmosContract)
			Expect(err).ToNot(HaveOccurred(), "failed to query ERC20 balance")
			Expect(balance.Int64()).To(Equal(int64(0)), "expected empty WEVMOS balance")
		})
	})
})
