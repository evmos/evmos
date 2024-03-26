package fixes

func AddMissingWalletsTetherA(wallets []string) []string {
	wallets = append(wallets, "0xce9ebc509f14cef1a37baaaa0e37da62568fd335")
	wallets = append(wallets, "0x6d5a1bfb13861c8da5732cdd4f3c512e7ac85a95")
	wallets = append(wallets, "0x01f5b36bc7260ac11dfe192c2efc60bdeb5c8b63")
	wallets = append(wallets, "0xe2a36ceea3c16349ed991a52bcf5fa74f041c132")
	wallets = append(wallets, "0xb804224f45da11fdc2aed1b136b6aab24e3a416c")
	wallets = append(wallets, "0x1d5b33c83e284b7894c073fed950d95f8384dd05")
	wallets = append(wallets, "0x8303b802a00f998b21a8bcb48c195121e26b6e6b")
	wallets = append(wallets, "0xeeacc00244feff0d68a4773547d05c1e24edee7e")
	wallets = append(wallets, "0xae49a8b0189521b3c6d0cd7fa62ec2f9036b2d74")
	return wallets
}
