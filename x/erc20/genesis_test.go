package erc20_test

import (
	"github.com/stretchr/testify/suite"
)


type GenesisTestSuite struct {
	suite.Suite
	ctx sdk.Context
	app     *app.Evmos
	genesis types.GenesisState
}

func (suite *GenesisTestSuite) SetupTest(){
	
}

func (suite *GenesisTestSuite) TestERC20InitGenesis(t *testing.T){

}

func (suite *GenesisTestSuite) TestErc20ExportGenesis(t *testing.T){
	
}