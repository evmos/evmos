package backend

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	tmrpcclient "github.com/cometbft/cometbft/rpc/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/evmos/v20/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v20/rpc/backend/mocks"
	"github.com/evmos/evmos/v20/server/config"
	"github.com/evmos/evmos/v20/types"
	evmconfig "github.com/evmos/evmos/v20/x/evm/config"
	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
)

func (suite *BackendTestSuite) TestRPCMinGasPrice() {
	testCases := []struct {
		name           string
		registerMock   func()
		expMinGasPrice int64
		expPass        bool
	}{
		{
			"pass - default gas price",
			func() {
			},
			types.DefaultGasPrice,
			true,
		},
		{
			"pass - min gas price is 0",
			func() {
			},
			types.DefaultGasPrice,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			minPrice := suite.backend.RPCMinGasPrice()
			if tc.expPass {
				suite.Require().Equal(tc.expMinGasPrice, minPrice)
			} else {
				suite.Require().NotEqual(tc.expMinGasPrice, minPrice)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGenerateMinGasCoin() {
	defaultGasPrice := (*hexutil.Big)(big.NewInt(1))
	testCases := []struct {
		name           string
		gasPrice       hexutil.Big
		minGas         sdk.DecCoins
		expectedOutput sdk.DecCoin
	}{
		{
			"pass - empty min gas Coins (default denom)",
			*defaultGasPrice,
			sdk.DecCoins{},
			sdk.DecCoin{
				Denom:  evmconfig.GetEVMCoinDenom(),
				Amount: math.LegacyNewDecFromBigInt(defaultGasPrice.ToInt()),
			},
		},
		{
			"pass - different min gas Coin",
			*defaultGasPrice,
			sdk.DecCoins{sdk.NewDecCoin("test", math.NewInt(1))},
			sdk.DecCoin{
				Denom:  "test",
				Amount: math.LegacyNewDecFromBigInt(defaultGasPrice.ToInt()),
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			suite.backend.clientCtx.Viper = viper.New()

			appConf := config.DefaultConfig()
			appConf.SetMinGasPrices(tc.minGas)

			output := suite.backend.GenerateMinGasCoin(tc.gasPrice, *appConf)
			suite.Require().Equal(tc.expectedOutput, output)
		})
	}
}

// TODO: Combine these 2 into one test since the code is identical
func (suite *BackendTestSuite) TestListAccounts() {
	testCases := []struct {
		name         string
		registerMock func()
		expAddr      []common.Address
		expPass      bool
	}{
		{
			"pass - returns empty address",
			func() {},
			[]common.Address{},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			output, err := suite.backend.ListAccounts()

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expAddr, output)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestAccounts() {
	testCases := []struct {
		name         string
		registerMock func()
		expAddr      []common.Address
		expPass      bool
	}{
		{
			"pass - returns empty address",
			func() {},
			[]common.Address{},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			output, err := suite.backend.Accounts()

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expAddr, output)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestSyncing() {
	testCases := []struct {
		name         string
		registerMock func()
		expResponse  interface{}
		expPass      bool
	}{
		{
			"fail - Can't get status",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterStatusError(client)
			},
			false,
			false,
		},
		{
			"pass - Node not catching up",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterStatus(client)
			},
			false,
			true,
		},
		{
			"pass - Node is catching up",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterStatus(client)
				status, _ := client.Status(suite.backend.ctx)
				status.SyncInfo.CatchingUp = true
			},
			map[string]interface{}{
				"startingBlock": hexutil.Uint64(0),
				"currentBlock":  hexutil.Uint64(0),
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			output, err := suite.backend.Syncing()

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expResponse, output)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestSetEtherbase() {
	testCases := []struct {
		name         string
		registerMock func()
		etherbase    common.Address
		expResult    bool
	}{
		{
			"pass - Failed to get coinbase address",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterStatusError(client)
			},
			common.Address{},
			false,
		},
		{
			"pass - the minimum fee is not set",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterStatus(client)
				RegisterValidatorAccount(queryClient, suite.acc)
			},
			common.Address{},
			false,
		},
		{
			"fail - error querying for account",
			func() {
				var header metadata.MD
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterStatus(client)
				RegisterValidatorAccount(queryClient, suite.acc)
				RegisterParams(queryClient, &header, 1)
				c := sdk.NewDecCoin(types.BaseDenom, math.NewIntFromBigInt(big.NewInt(1)))
				suite.backend.cfg.SetMinGasPrices(sdk.DecCoins{c})
				delAddr, _ := suite.backend.GetCoinbase()
				// account, _ := suite.backend.clientCtx.AccountRetriever.GetAccount(suite.backend.clientCtx, delAddr)
				delCommonAddr := common.BytesToAddress(delAddr.Bytes())
				request := &authtypes.QueryAccountRequest{Address: sdk.AccAddress(delCommonAddr.Bytes()).String()}
				requestMarshal, _ := request.Marshal()
				RegisterABCIQueryWithOptionsError(
					client,
					"/cosmos.auth.v1beta1.Query/Account",
					requestMarshal,
					tmrpcclient.ABCIQueryOptions{Height: int64(1), Prove: false},
				)
			},
			common.Address{},
			false,
		},
		// TODO: Finish this test case once ABCIQuery GetAccount is fixed
		// {
		//	"pass - set the etherbase for the miner",
		//	func() {
		//		client := suite.backend.clientCtx.Client.(*mocks.Client)
		//		queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
		//		RegisterStatus(client)
		//		RegisterValidatorAccount(queryClient, suite.acc)
		//		c := sdk.NewDecCoin(types.AttoEvmos, math.NewIntFromBigInt(big.NewInt(1)))
		//		suite.backend.cfg.SetMinGasPrices(sdk.DecCoins{c})
		//		delAddr, _ := suite.backend.GetCoinbase()
		//		account, _ := suite.backend.clientCtx.AccountRetriever.GetAccount(suite.backend.clientCtx, delAddr)
		//		delCommonAddr := common.BytesToAddress(delAddr.Bytes())
		//		request := &authtypes.QueryAccountRequest{Address: sdk.AccAddress(delCommonAddr.Bytes()).String()}
		//		requestMarshal, _ := request.Marshal()
		//		RegisterABCIQueryAccount(
		//			client,
		//			requestMarshal,
		//			tmrpcclient.ABCIQueryOptions{Height: int64(1), Prove: false},
		//			account,
		//		)
		//	},
		//	common.Address{},
		//	false,
		// },
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			output := suite.backend.SetEtherbase(tc.etherbase)

			suite.Require().Equal(tc.expResult, output)
		})
	}
}

func (suite *BackendTestSuite) TestImportRawKey() {
	priv, _ := ethsecp256k1.GenerateKey()
	privHex := common.Bytes2Hex(priv.Bytes())
	pubAddr := common.BytesToAddress(priv.PubKey().Address().Bytes())

	testCases := []struct {
		name         string
		registerMock func()
		privKey      string
		password     string
		expAddr      common.Address
		expPass      bool
	}{
		{
			"fail - not a valid private key",
			func() {},
			"",
			"",
			common.Address{},
			false,
		},
		{
			"pass - returning correct address",
			func() {},
			privHex,
			"",
			pubAddr,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			output, err := suite.backend.ImportRawKey(tc.privKey, tc.password)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expAddr, output)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
