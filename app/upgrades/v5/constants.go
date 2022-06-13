package v5

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// UpgradeName is the shared upgrade plan name for mainnet and testnet
	UpgradeName = "v5.0.0"
	// MainnetUpgradeHeight defines the Evmos mainnet block height on which the upgrade will take place
	// TODO: define
	MainnetUpgradeHeight = 257_850
	// TestnetUpgradeHeight defines the Evmos testnet block height on which the upgrade will take place
	// TODO: define
	TestnetUpgradeHeight = 1_200_000
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/arm64":"https://github.com/tharsis/evmos/releases/download/v5.0.0/evmos_5.0.0_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/tharsis/evmos/releases/download/v5.0.0/evmos_5.0.0_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/tharsis/evmos/releases/download/v5.0.0/evmos_5.0.0_Linux_arm64.tar.gz","linux/x86_64":"https://github.com/tharsis/evmos/releases/download/v5.0.0/evmos_5.0.0_Linux_x86_64.tar.gz","windows/x86_64":"https://github.com/tharsis/evmos/releases/download/v5.0.0/evmos_5.0.0_Windows_x86_64.zip"}}'`
	// ContributorAddrFrom is the lost address of an early contributor
	ContributorAddrFrom = "evmos13cf9npvns2vhh3097909mkhfxngmw6d6eppfm4"
	// ContributorAddrTo is the new address of an early contributor
	ContributorAddrTo = "evmos1hmntpkn623y3vl0nvzrazvq4rqzv3xa74l40gl"
)

var (
	// MainnetMinGasPrices defines 25B aevmos (or atevmos) as the minimum gas price value on the fee market module.
	// See https://commonwealth.im/evmos/discussion/5073-global-min-gas-price-value-for-cosmos-sdk-and-evm-transaction-choosing-a-value for reference
	MainnetMinGasPrices = sdk.NewDec(25_000_000_000)
	// MainnetMinGasMultiplier defines the min gas multiplier value on the fee market module.
	// 50% of the leftover gas will be refunded
	MainnetMinGasMultiplier = sdk.NewDecWithPrec(5, 1)
)
