package integration_test_util

//goland:noinspection SpellCheckingInspection
import (
	"fmt"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	itutiltypes "github.com/evmos/evmos/v16/integration_test_util/types"
	erc20types "github.com/evmos/evmos/v16/x/erc20/types"
	"strings"
)

// TxFullRegisterCoin registers a coin with the given denom and amount using Governance module.
// Auto propose a new proposal, vote and wait till it done.
func (suite *ChainIntegrationTestSuite) TxFullRegisterCoin(proposer *itutiltypes.TestAccount, minDenom string) uint64 {
	suite.Require().NotNil(proposer)

	denomRes, err := suite.QueryClients.Bank.DenomMetadata(suite.CurrentContext, &banktypes.QueryDenomMetadataRequest{
		Denom: minDenom,
	})
	suite.Require().NoError(err)
	suite.Require().NotNilf(denomRes, "must register denom metadata for %s during integration setup for re-use purpose", minDenom)

	return suite.TxFullRegisterCoinByMetadata(proposer, denomRes.Metadata)
}

// TxFullRegisterCoinWithNewBankMetadata registers a coin with the given denom metadata using Governance module.
// Auto propose a new proposal, vote and wait till it done.
func (suite *ChainIntegrationTestSuite) TxFullRegisterCoinWithNewBankMetadata(proposer *itutiltypes.TestAccount, minDenom, display string, exponent uint32) uint64 {
	suite.Require().NotNil(proposer)

	return suite.TxFullRegisterCoinByMetadata(proposer, banktypes.Metadata{
		Description: fmt.Sprintf("Denom metadata of %s", minDenom),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    minDenom,
				Exponent: 0,
			},
			{
				Denom:    display,
				Exponent: exponent,
			},
		},
		Base:    minDenom,
		Display: display,
		Name:    display,
		Symbol:  strings.ToUpper(display),
	})
}

// TxFullRegisterCoinByMetadata registers a coin with the given denom metadata using Governance module.
// Auto propose a new proposal, vote and wait till it done.
func (suite *ChainIntegrationTestSuite) TxFullRegisterCoinByMetadata(proposer *itutiltypes.TestAccount, metadata banktypes.Metadata) uint64 {
	suite.Require().NotNil(proposer)

	// TODO Fix me

	suite.T().Skip("TODO: fix me")
	//content := erc20types.NewRegisterCoinProposal(
	//	fmt.Sprintf("Register ERC-20 token pairs for %s", metadata.Base),
	//	fmt.Sprintf("Register ERC-20 token pairs for %s", metadata.Base),
	//	metadata,
	//)

	//return suite.TxFullGov(proposer, content)

	return 0
}

// TxFullRegisterIbcCoinFromErc20Contract registers IBC coin for the given ERC-20 contract address using Governance module.
// Auto propose a new proposal, vote and wait till it done.
func (suite *ChainIntegrationTestSuite) TxFullRegisterIbcCoinFromErc20Contract(proposer *itutiltypes.TestAccount, erc20Address common.Address) uint64 {
	suite.Require().NotNil(proposer)

	content := erc20types.NewRegisterERC20Proposal(
		fmt.Sprintf("Register IBC token pairs for ERC-20 contract %s", erc20Address.String()),
		fmt.Sprintf("Register IBC token pairs for ERC-20 contract %s", erc20Address.String()),
		erc20Address.String(),
	)

	return suite.TxFullGov(proposer, content)
}

// QueryFirstErc20TokenPair returns the first ERC-20 token pair available.
func (suite *ChainIntegrationTestSuite) QueryFirstErc20TokenPair(sourceErc20 bool) (erc20types.TokenPair, error) {
	tokenPairs, err := suite.QueryClients.Erc20.TokenPairs(suite.CurrentContext, &erc20types.QueryTokenPairsRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(tokenPairs)
	suite.Require().GreaterOrEqual(len(tokenPairs.TokenPairs), 1)

	for _, pair := range tokenPairs.TokenPairs {
		if sourceErc20 && strings.HasPrefix(pair.Denom, "erc20/") {
			return pair, nil
		} else if !sourceErc20 && strings.HasPrefix(pair.Denom, "ibc/") {
			return pair, nil
		}
	}

	return erc20types.TokenPair{}, fmt.Errorf("no token pair found for sourceErc20=%v", sourceErc20)
}
