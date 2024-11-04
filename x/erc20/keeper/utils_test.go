package keeper_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/contracts"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/x/erc20/keeper/testdata"
	"github.com/evmos/evmos/v20/x/erc20/types"
	evm "github.com/evmos/evmos/v20/x/evm/types"
)

// MintFeeCollector mints some coins to the fee collector address.
// Use this only for unit tests. For integration tests, you can use the
// mintFeeCollector flag to setup some balance on genesis
func (suite *KeeperTestSuite) MintFeeCollector(coins sdk.Coins) {
	err := suite.network.App.BankKeeper.MintCoins(suite.network.GetContext(), types.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.network.App.BankKeeper.SendCoinsFromModuleToModule(suite.network.GetContext(), types.ModuleName, authtypes.FeeCollectorName, coins)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) DeployContract(name, symbol string, decimals uint8) (common.Address, error) {
	addr, err := suite.factory.DeployContract(
		suite.keyring.GetPrivKey(0),
		evm.EvmTxArgs{},
		factory.ContractDeploymentData{
			Contract:        contracts.ERC20MinterBurnerDecimalsContract,
			ConstructorArgs: []interface{}{name, symbol, decimals},
		},
	)
	if err != nil {
		return common.Address{}, err
	}

	return addr, suite.network.NextBlock()
}

func (suite *KeeperTestSuite) DeployContractMaliciousDelayed() (common.Address, error) {
	maliciousDelayedContract, err := testdata.LoadMaliciousDelayedContract()
	suite.Require().NoError(err, "failed to load malicious delayed contract")

	addr, err := suite.factory.DeployContract(
		suite.keyring.GetPrivKey(0),
		evm.EvmTxArgs{},
		factory.ContractDeploymentData{
			Contract:        maliciousDelayedContract,
			ConstructorArgs: []interface{}{big.NewInt(1000000000000000000)},
		},
	)
	if err != nil {
		return common.Address{}, err
	}

	return addr, suite.network.NextBlock()
}

func (suite *KeeperTestSuite) DeployContractDirectBalanceManipulation() (common.Address, error) {
	balanceManipulationContract, err := testdata.LoadBalanceManipulationContract()
	suite.Require().NoError(err, "failed to load balance manipulation contract")

	addr, err := suite.factory.DeployContract(
		suite.keyring.GetPrivKey(0),
		evm.EvmTxArgs{},
		factory.ContractDeploymentData{
			Contract:        balanceManipulationContract,
			ConstructorArgs: []interface{}{big.NewInt(1000000000000000000)},
		},
	)
	if err != nil {
		return common.Address{}, err
	}

	return addr, suite.network.NextBlock()
}
