package v17_test

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	testkeyring "github.com/evmos/evmos/v16/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/utils"
	"math/rand"
)

// nTokenPairs is the number of token pairs to generate
const nTokenPairs = 10

// CoinDenoms is a list of coin denoms used in the tests
var CoinDenoms = generateCoinDenoms(nTokenPairs)

func generateCoinDenoms(n int) []string {
	coinDenoms := make([]string, n)
	for i := 0; i < n; i++ {
		coinDenoms[i] = fmt.Sprintf("denom%02d", i)
	}

	return coinDenoms
}

func generateRandomBooleans(n int) []bool {
	randomBooleans := make([]bool, n)
	for i := 0; i < n; i++ {
		// generate random number between 0 and 1
		randNum := rand.Float32()
		if randNum < 0.5 {
			randomBooleans[i] = false
		} else {
			randomBooleans[i] = true
		}
	}

	return randomBooleans
}

func createRandomCoinBalances(n int) sdk.Coins {
	randomBools := generateRandomBooleans(n)
	coins := make(sdk.Coins, n)
	for i := 0; i < n; i++ {
		balance := sdkmath.ZeroInt()
		if randomBools[i] {
			balance = sdkmath.NewInt(int64(rand.Intn(1e18)))
		}
		coins[i] = sdk.NewCoin(CoinDenoms[i], balance)
	}

	return coins.Sort()
}

func createGenesisBalances(keyring testkeyring.Keyring) []banktypes.Balance {
	keys := keyring.GetAllAccAddrs()
	genesisBalances := make([]banktypes.Balance, len(keys))

	for i, acc := range keys {
		randomCoins := createRandomCoinBalances(nTokenPairs)
		genesisBalances[i] = banktypes.Balance{
			Address: acc.String(),
			Coins: randomCoins.Add(
				sdk.Coin{
					Denom:  utils.BaseDenom,
					Amount: network.PrefundedAccountInitialBalance,
				},
			),
		}
	}

	return genesisBalances
}

//// TODO: really necessary?
//func createCustomGenesis(keyring testkeyring.Keyring) (genesisState network.CustomGenesisState) {
//	genesisState = network.CustomGenesisState{}
//
//	bankGenesisState := network.BankCustomGenesisState{}
//	genesisState[banktypes.ModuleName] = bankGenesisState
//
//	return genesisState
//}
