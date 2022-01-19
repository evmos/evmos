package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/osmosis-labs/osmosis/x/mint/types"
)

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	// minter

	// var maxRewardPerEpoch sdk.Dec
	// simState.AppParams.GetOrGenerate(
	// 	simState.Cdc, MaxRewardPerEpoch, &maxRewardPerEpoch, simState.Rand,
	// 	func(r *rand.Rand) { maxRewardPerEpoch = GenMaxRewardPerEpoch(r) },
	// )

	// var minRewardPerEpoch sdk.Dec
	// simState.AppParams.GetOrGenerate(
	// 	simState.Cdc, MinRewardPerEpoch, &minRewardPerEpoch, simState.Rand,
	// 	func(r *rand.Rand) { minRewardPerEpoch = GenMinRewardPerEpoch(r) },
	// )
	// Leaving as sample code

	mintDenom := sdk.DefaultBondDenom
	epochProvisions := sdk.NewDec(500000) // TODO: Randomize this
	params := types.NewParams(mintDenom, epochProvisions, "week", sdk.NewDecWithPrec(5, 1), 156, types.DistributionProportions{
		Staking:          sdk.NewDecWithPrec(4, 1), // 0.4
		PoolIncentives:   sdk.NewDecWithPrec(3, 1), // 0.3
		DeveloperRewards: sdk.NewDecWithPrec(2, 1), // 0.2
		CommunityPool:    sdk.NewDecWithPrec(1, 1), // 0.1
	}, []types.WeightedAddress{}, 0)

	mintGenesis := types.NewGenesisState(types.InitialMinter(), params, 0)

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	// TODO: Do some randomization later
	fmt.Printf("Selected deterministically generated minting parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}
