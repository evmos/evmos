package fixes

import (
	"slices"
"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/ethereum/go-ethereum/common"
)

func GetAllMissingWallets() []string {
	var wallets []string
	// DAI-AXELAR
	wallets = append(wallets, "0xf3c623b226331ee73b14e7b2b2025bd8be85db92")

	// TORI
	wallets = append(wallets, "0xfd729ccb71863f3ff7e2aa92acb6940a21d64911")
	wallets = append(wallets, "0x73668ef732b66f1841c8b8ef4fd16bb9570ef038")

	// WETH-AXELAR
	wallets = append(wallets, "0xf1f007c47d5afb09d327c1edc8c9c63a027f6f66")

	// Wormhole Bitcoin
	wallets = append(wallets, "0x440a1ef3c8e25475f4f3d546172ecaca0ddf62bc")

	// Axelar
	wallets = append(wallets, "0x3edcdf87403200b14ad117b24b9de6777751d3a4")
	wallets = append(wallets, "0xaec7342c2cf450f5538e91125566f2a77a851595")
	wallets = append(wallets, "0x6433a9e9834e004cf129a73992e2ca70625162f2")
	wallets = append(wallets, "0xdee2a06e21336152d089ced6699efb0d7bdc0dc3")
	wallets = append(wallets, "0x4c42bd1f74b3e93873c54d8116571ec33184704b")
	wallets = append(wallets, "0x34b5eba8cfc392634ec292886121c23c2cc420fe")

	// USDC Axelar
	wallets = AddMissingWalletsUSDCA(wallets)

	// Graivity USDC
	wallets = AddMissingWalletsUSDCG(wallets)

	// WETH G
	wallets = AddMissingWalletsWETHG(wallets)

	// Graviton
	wallets = AddMissingWalletsGraviton(wallets)

	// USDT Kava
	wallets = append(wallets, "0x75656beebb6213709bab9ecb2a9f5191a781e3e5")
	wallets = append(wallets, "0xb97ee67e5f3fd3e2c22ded8ed6ca19693181d72d")
	wallets = append(wallets, "0xbb4d2dcaa5412040c59de4a1ff576289da83f9a3")
	wallets = append(wallets, "0xc71c9bed735b6e6423624b8c17099e553367a474")
	wallets = append(wallets, "0xe0568365a57a08cf066fc00e4512cdf7b0fe87e0")
	wallets = append(wallets, "0xd499b248b8ee592b424fbbeac115b9a5733e2508")
	wallets = append(wallets, "0x23c1e98ed321b41ab6592ef6f08d2f506d02b1a4")
	wallets = append(wallets, "0x77ee032a4c0dca739512191e2f2a4fbd54738853")
	wallets = append(wallets, "0x18233f548849786d7e467ca0fe7f9863e61bdfc4")
	wallets = append(wallets, "0x22922f2086edd7b2ee47673409c4c1dd3403a503")
	wallets = append(wallets, "0x67c301eda4e11cce806cf0cda323aa556004b851")
	wallets = append(wallets, "0x028f17575aa7f6b98da6661473b9488d6ce9dddb")
	wallets = append(wallets, "0x028f17575aa7f6b98da6661473b9488d6ce9dddb")
	wallets = append(wallets, "0x2406690c1513b0e9f6fb5084d8c956b0d9fc6d08")
	wallets = append(wallets, "0x2406690c1513b0e9f6fb5084d8c956b0d9fc6d08")
	wallets = append(wallets, "0x9c0f498d69df554682e8bd8c0ec1d71fac76ec91")
	wallets = append(wallets, "0x272c35eed0e96c0bf05d0d4071545e1e92d5206a")
	wallets = append(wallets, "0x02899e083dcb403ba1a1e325c95551c5fe0da817")
	wallets = append(wallets, "0x6a31c8b9a783aece03426efb5b29984fea16cb3c")
	wallets = append(wallets, "0x7c7f3837d43d9af9a2a3b5ac082178173bd4d708")
	wallets = append(wallets, "0x57d539ad75e1ae12ece5a405d898916a6fb68faa")
	wallets = append(wallets, "0xca282181defb61269dbbfb183c554a7d4a4a204e")
	wallets = append(wallets, "0xc3ef6458b92a6ebfade52b74e451c861cce73b9c")
	wallets = append(wallets, "0x769249dce2065e677699a94b9c0bb46f079a90ac")
	wallets = append(wallets, "0x19fb3225b9512090a3b2d35cc516af77249fb650")
	wallets = append(wallets, "0x9c0f498d69df554682e8bd8c0ec1d71fac76ec91")
	wallets = append(wallets, "0xafc86dd630a8814f9650cd2312217eb4617fb9b7")
	wallets = append(wallets, "0x0c0d535f1683e78dcacf1a7702132853ec9d7d67")
	wallets = append(wallets, "0x14cba0298c152d9cd1cd70008a45dbf3708eaafa")
	wallets = append(wallets, "0x5e5577032c0aa883d8ea401d64606c07dacaf338")
	wallets = append(wallets, "0x4cff6ac23f7b51b0569bc6b8abca43f98f74cbfc")
	wallets = append(wallets, "0xb6e2be59dff3b394ad77508527f375a257a4738e")
	wallets = append(wallets, "0x18233f548849786d7e467ca0fe7f9863e61bdfc4")
	wallets = append(wallets, "0x6a1d542fe756394fb27b724d22dc1b288ad241c6")
	wallets = append(wallets, "0x4c3914ce2381d1f8cafb4efc363325fd54ee5ee3")
	wallets = append(wallets, "0xaf06613e7d5e58458e647d9da6780cf481887900")
	wallets = append(wallets, "0x886435885f29353c5b32583b50ebe4233ada0808")
	wallets = append(wallets, "0xc00c068d8f98997ecb8f2080f427c4c9466b9317")
	wallets = append(wallets, "0x1afd31627170607657224ed4ae701470209c4b2e")
	wallets = append(wallets, "0x0c1d72fbb8b8c19440c4d96b3b1b2e8d4aa5dfb4")
	wallets = append(wallets, "0x028f17575aa7f6b98da6661473b9488d6ce9dddb")
	wallets = append(wallets, "0x67c301eda4e11cce806cf0cda323aa556004b851")
	wallets = append(wallets, "0x22922f2086edd7b2ee47673409c4c1dd3403a503")
	wallets = append(wallets, "0x6302f228a9cecbd771d1875060d378cc1797487a")
	wallets = append(wallets, "0x235177452a2b37d2770c35ff9b669a8fe894accd")
	wallets = append(wallets, "0x2406690c1513b0e9f6fb5084d8c956b0d9fc6d08")

	// ATOM
	wallets = AddMissingWalletsATOM(wallets)
	// !IMPORTANT: transaction found using the finder binary
	wallets = append(wallets, "0x22cbe6caf10f40bc143a7762e2b4c007c95f6c16")

	// StEvmos
	wallets = AddMissingWalletsStEvmos(wallets)

	// TetherG
	wallets = AddMissingWalletsTetherG(wallets)

	// Osmosis
	wallets = AddMissingWalletsOsmosis(wallets)

	// TetherA
	wallets = AddMissingWalletsTetherA(wallets)

	// DaiG
	wallets = AddMissingWalletsDaiG(wallets)

	// Unique filter
	slices.Sort(wallets)
	wallets = slices.Compact(wallets)
	return wallets
}

func GetMissingWalletsFromAuthModule(ctx sdk.Context,
	accountKeeper authkeeper.AccountKeeper) (Addresses []sdk.AccAddress) {
	wallets := GetAllMissingWallets()
	for _, wallet := range wallets {
		ethAddr := common.HexToAddress(wallet)
		addr := sdk.AccAddress(ethAddr.Bytes())
		if accountKeeper.HasAccount(ctx, addr) {
			fmt.Println("Account existed")
			continue
		}
		Addresses = append(Addresses, addr)
	}

	return Addresses
}
