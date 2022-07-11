package v7

const (
	// UpgradeName is the shared upgrade plan name for mainnet and testnet
	UpgradeName = "v7.0.0"
	// TODO MainnetUpgradeHeight defines the Evmos mainnet block height on which the upgrade will take place
	MainnetUpgradeHeight = 1_042_000
	// TODO TestnetUpgradeHeight defines the Evmos testnet block height on which the upgrade will take place
	TestnetUpgradeHeight = 2_176_500
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/arm64":"https://github.com/evmos/evmos/releases/download/v7.0.0/evmos_7.0.0_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/evmos/evmos/releases/download/v7.0.0/evmos_7.0.0_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/evmos/evmos/releases/download/v7.0.0/evmos_7.0.0_Linux_arm64.tar.gz","linux/x86_64":"https://github.com/evmos/evmos/releases/download/v7.0.0/evmos_7.0.0_Linux_x86_64.tar.gz","windows/x86_64":"https://github.com/evmos/evmos/releases/download/v7.0.0/evmos_7.0.0_Windows_x86_64.zip"}}'`

	// FaucetAddressFrom is the inaccessible secp address of the Testnet Faucet
	FaucetAddressFrom = "evmos1z4ya98ga2xnffn2mhjym7tzlsm49ec23890sze"
	// FaucetAddressTo is the new eth_secp address of the Testnet Faucet
	FaucetAddressTo = "evmos1ujm4z5v9zkdqm70xnptr027gqu90f7lxjr0fch"
)
