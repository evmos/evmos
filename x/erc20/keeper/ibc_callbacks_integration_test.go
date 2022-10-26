package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	
	ibcgotesting "github.com/cosmos/ibc-go/v5/testing"
	ibctesting "github.com/evmos/evmos/v9/ibc/testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	claimtypes "github.com/evmos/evmos/v9/x/claims/types"
	inflationtypes "github.com/evmos/evmos/v9/x/inflation/types"
	recoverytypes "github.com/evmos/evmos/v9/x/recovery/types"
	"github.com/evmos/evmos/v9/x/erc20/types"

	//"github.com/evmos/evmos/v9/x/erc20/keeper"
	"github.com/evmos/evmos/v9/app"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	"github.com/cosmos/cosmos-sdk/baseapp"
	evm "github.com/evmos/ethermint/x/evm/types"
	ethermint "github.com/evmos/ethermint/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	//"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/evmos/ethermint/tests"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/encoding"
	"github.com/cosmos/cosmos-sdk/client"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// VARS
var (

	uosmoDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uosmo",
	}

	uosmoIbcdenom = uosmoDenomtrace.IBCDenom()

	uatomDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-1",
		BaseDenom: "uatom",
	}
	uatomIbcdenom = uatomDenomtrace.IBCDenom()

	aevmosDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "aevmos",
	}
	aevmosIbcdenom = aevmosDenomtrace.IBCDenom()

	uatomOsmoDenomtrace = transfertypes.DenomTrace{
		Path:      "transfer/channel-0/transfer/channel-1",
		BaseDenom: "uatom",
	}
	uatomOsmoIbcdenom = uatomOsmoDenomtrace.IBCDenom()
)

// IBC TESTS
func (suite *KeeperTestSuite) TestIBCIntegration() {

	// Initializes 3 test chains
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 1, 2)
	suite.EvmosChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(1))
	suite.IBCOsmosisChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(2))
	suite.IBCCosmosChain = suite.coordinator.GetChain(ibcgotesting.GetChainID(3))
	suite.coordinator.CommitNBlocks(suite.EvmosChain, 2)
	suite.coordinator.CommitNBlocks(suite.IBCOsmosisChain, 2)
	suite.coordinator.CommitNBlocks(suite.IBCCosmosChain, 2)

	// Mint coins locked on the evmos account generated with secp.
	coinEvmos := sdk.NewCoin("aevmos", sdk.NewInt(1000))
	coins := sdk.NewCoins(coinEvmos)
	err := suite.EvmosChain.App.(*app.Evmos).BankKeeper.MintCoins(suite.EvmosChain.GetContext(), inflationtypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.EvmosChain.App.(*app.Evmos).BankKeeper.SendCoinsFromModuleToAccount(suite.EvmosChain.GetContext(), inflationtypes.ModuleName, suite.EvmosChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	// Mint coins on the osmosis side which we'll send over to evmos (won't be converted, not creating token pair)
	coinOsmo := sdk.NewCoin("uosmo", sdk.NewInt(1000))
	coins = sdk.NewCoins(coinOsmo)
	err = suite.IBCOsmosisChain.GetSimApp().BankKeeper.MintCoins(suite.IBCOsmosisChain.GetContext(), minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.IBCOsmosisChain.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(suite.IBCOsmosisChain.GetContext(), minttypes.ModuleName, suite.IBCOsmosisChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	// Mint coins on the cosmos side which we'll send over to evmos (will be converted, creating token pair)
	coinAtom := sdk.NewCoin("uatom", sdk.NewInt(1000))
	coins = sdk.NewCoins(coinAtom)
	err = suite.IBCCosmosChain.GetSimApp().BankKeeper.MintCoins(suite.IBCCosmosChain.GetContext(), minttypes.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.IBCCosmosChain.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(suite.IBCCosmosChain.GetContext(), minttypes.ModuleName, suite.IBCCosmosChain.SenderAccount.GetAddress(), coins)
	suite.Require().NoError(err)

	// Create paths
	suite.pathOsmosisEvmos = ibctesting.NewTransferPath(suite.IBCOsmosisChain, suite.EvmosChain) // clientID, connectionID, channelID empty
	suite.pathCosmosEvmos = ibctesting.NewTransferPath(suite.IBCCosmosChain, suite.EvmosChain)
	suite.pathOsmosisCosmos = ibctesting.NewTransferPath(suite.IBCCosmosChain, suite.IBCOsmosisChain)
	suite.coordinator.Setup(suite.pathOsmosisEvmos) // clientID, connectionID, channelID filled
	suite.coordinator.Setup(suite.pathCosmosEvmos)
	suite.coordinator.Setup(suite.pathOsmosisCosmos)
	suite.Require().Equal("07-tendermint-0", suite.pathOsmosisEvmos.EndpointA.ClientID)
	suite.Require().Equal("connection-0", suite.pathOsmosisEvmos.EndpointA.ConnectionID)
	suite.Require().Equal("channel-0", suite.pathOsmosisEvmos.EndpointA.ChannelID)

	// Set up Evmos Chain w/ EVM, ERC20 Module
	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = tests.NewSigner(priv)

	priv, err = ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// Important: controls context, allows us to make ERC-20 Keeper calls
	suite.EvmosChain.CurrentHeader = tmproto.Header{
		Height:          suite.EvmosChain.CurrentHeader.Height, // starting point considering IBC
		ChainID:         "evmos_9001-1",
		Time:            time.Now().UTC(),
		ProposerAddress: suite.consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	}

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.EvmosChain.GetContext(), suite.EvmosChain.App.(*app.Evmos).InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, suite.EvmosChain.App.(*app.Evmos).EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelperEvm)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.EvmosChain.GetContext(), suite.EvmosChain.App.(*app.Evmos).InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.EvmosChain.App.(*app.Evmos).Erc20Keeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	ethacc := &ethermint.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.EvmosChain.App.(*app.Evmos).AccountKeeper.SetAccount(suite.EvmosChain.GetContext(), ethacc)

	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	suite.Require().NoError(err)
	err = suite.EvmosChain.App.(*app.Evmos).StakingKeeper.SetValidatorByConsAddr(suite.EvmosChain.GetContext(), validator)
	suite.Require().NoError(err)
	suite.EvmosChain.App.(*app.Evmos).StakingKeeper.SetValidator(suite.EvmosChain.GetContext(), validator)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.EvmosChain.App.(*app.Evmos).EvmKeeper.ChainID())

	// Set params for claims, recovery, and ERC-20 module
	claimparams := claimtypes.DefaultParams()
	claimparams.AirdropStartTime = suite.EvmosChain.GetContext().BlockTime()
	claimparams.EnableClaims = false // claims complete at the time of adding ERC20 IBC Middleware
	suite.EvmosChain.App.(*app.Evmos).ClaimsKeeper.SetParams(suite.EvmosChain.GetContext(), claimparams)

	recoveryparams := recoverytypes.DefaultParams()
	recoveryparams.EnableRecovery = true
	suite.EvmosChain.App.(*app.Evmos).RecoveryKeeper.SetParams(suite.EvmosChain.GetContext(), recoveryparams)

	erc20params := types.DefaultParams()
	erc20params.EnableErc20 = true
	suite.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(suite.EvmosChain.GetContext(), erc20params)

	// Register ATOM with a Token Pair for testing
	validMetadata := banktypes.Metadata{
		Description: "IBC Coin for IBC Cosmos Chain",
		Base:        uatomDenomtrace.BaseDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    uatomDenomtrace.BaseDenom,
				Exponent: 0,
			},
		},
		Name:    uatomDenomtrace.BaseDenom,
		Symbol:  erc20Symbol,
		Display: uatomDenomtrace.BaseDenom,
	}

	err = suite.EvmosChain.App.(*app.Evmos).BankKeeper.MintCoins(suite.EvmosChain.GetContext(), inflationtypes.ModuleName, sdk.Coins{sdk.NewInt64Coin(validMetadata.Base, 1)})
	suite.Require().NoError(err)

	_, err = suite.EvmosChain.App.(*app.Evmos).Erc20Keeper.RegisterCoin(suite.EvmosChain.GetContext(), validMetadata)
    suite.Require().NoError(err)

	// Set up packet
	timeoutHeight := clienttypes.NewHeight(1, 1000)
	disabledTimeoutTimestamp := uint64(0)
	mockPacket := channeltypes.NewPacket(ibcgotesting.MockPacketData, 1, transfertypes.PortID, "channel-0", transfertypes.PortID, "channel-0", timeoutHeight, disabledTimeoutTimestamp)
	packet := mockPacket

	// Tests:
	// - don't need to check non ics-20 packets, blocked sender/recipient, handled at deeper level
	// - need to check disabled ERC-20s
	// - need to check sending a coin w/ token pair across IBC
	// - need to check sending a coin w/o token pair across IBC
	// - need to check sending aevmos across IBC
	// - need to check sending ERC-20s across IBC
	testCases := []struct {
		name         string
		malleate     func()
		currentPath  *ibcgotesting.Path
	}{
		{
			"no-op: erc-20 disabled",
			func() {
				erc20params := types.DefaultParams()
				erc20params.EnableErc20 = false
				suite.EvmosChain.App.(*app.Evmos).Erc20Keeper.SetParams(suite.EvmosChain.GetContext(), erc20params)

				// create correct packet
				path := suite.pathCosmosEvmos
				transfer := transfertypes.NewFungibleTokenPacketData("uatom", "100", suite.IBCCosmosChain.SenderAccount.GetAddress().String(), suite.EvmosChain.SenderAccount.GetAddress().String())
				bz := transfertypes.ModuleCdc.MustMarshalJSON(&transfer)
				packet = channeltypes.NewPacket(bz, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, 0)
			},
			suite.pathCosmosEvmos,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			// Set path and malleate function
			path := tc.currentPath
			tc.malleate()

			// send on endpointA
			path.EndpointA.SendPacket(packet)

			// receive on endpointB
			_ = path.EndpointB.RecvPacket(packet)
			// suite.Require().Error(err)
		})
	}
}