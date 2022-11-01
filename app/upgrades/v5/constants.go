package v5

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// UpgradeName is the shared upgrade plan name for mainnet and testnet
	UpgradeName = "v5.0.0"
	// TestnetUpgradeHeight defines the Evoblock testnet block height on which the upgrade will take place
	TestnetUpgradeHeight = 1_762_500
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/arm64":"https://github.com/evoblockchain/evoblock/releases/download/v5.0.0/evoblock_5.0.0_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/evoblockchain/evoblock/releases/download/v5.0.0/evoblock_5.0.0_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/evoblockchain/evoblock/releases/download/v5.0.0/evoblock_5.0.0_Linux_arm64.tar.gz","linux/x86_64":"https://github.com/evoblockchain/evoblock/releases/download/v5.0.0/evoblock_5.0.0_Linux_amd64.tar.gz","windows/x86_64":"https://github.com/evoblockchain/evoblock/releases/download/v5.0.0/evoblock_5.0.0_Windows_x86_64.zip"}}'`
	// ContributorAddrFrom is the lost address of an early contributor
	ContributorAddrFrom = "evo13cf9npvns2vhh3097909mkhfxngmw6d6eppfm4"
	// ContributorAddrTo is the new address of an early contributor
	ContributorAddrTo = "evo1hmntpkn623y3vl0nvzrazvq4rqzv3xa74l40gl"
	// AvgBlockTime defines the new expected average blocktime on mainnet and testnet
	//
	// CONTRACT: in order for AvgBlockTime to represent an accurate value on-chain, validator nodes
	// will need to update their "timeout_commit" value to "1s" on the config.toml under the
	// "Consensus Configuration Options" section
	//
	// NOTE: the value is calculated based that it takes <1s to reach consensus and 1s for
	// "timeout_commit" duration
	AvgBlockTime = 2 * time.Second
)

var (
	// MainnetMinGasPrices defines 20B aEVO (or atEVO) as the minimum gas price value on the fee market module.
	// See https://commonwealth.im/evoblock/discussion/5073-global-min-gas-price-value-for-cosmos-sdk-and-evm-transaction-choosing-a-value for reference
	MainnetMinGasPrices = sdk.NewDec(20_000_000_000)
	// MainnetMinGasMultiplier defines the min gas multiplier value on the fee market module.
	// 50% of the leftover gas will be refunded
	MainnetMinGasMultiplier = sdk.NewDecWithPrec(5, 1)
)
