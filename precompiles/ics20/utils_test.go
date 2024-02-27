// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package ics20_test

import (
	"cosmossdk.io/math"
	"fmt"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	evmosapp "github.com/evmos/evmos/v16/app"
	"github.com/evmos/evmos/v16/precompiles/authorization"
	cmn "github.com/evmos/evmos/v16/precompiles/common"
	"github.com/evmos/evmos/v16/precompiles/ics20"
	evmosutil "github.com/evmos/evmos/v16/testutil"
	commonnetwork "github.com/evmos/evmos/v16/testutil/integration/common/network"
	"github.com/evmos/evmos/v16/testutil/integration/ibc/coordinator"
	evmosutiltx "github.com/evmos/evmos/v16/testutil/tx"
	"github.com/evmos/evmos/v16/utils"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
	"math/big"
	//nolint:revive // dot imports are fine for Ginkgo
)

type erc20Meta struct {
	Name     string
	Symbol   string
	Decimals uint8
}

var (
	maxUint256Coins    = sdk.Coins{sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewIntFromBigInt(abi.MaxUint256)}}
	maxUint256CmnCoins = []cmn.Coin{{Denom: utils.BaseDenom, Amount: abi.MaxUint256}}
	defaultCoins       = sdk.Coins{sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)}}
	baseDenomCmnCoin   = cmn.Coin{Denom: utils.BaseDenom, Amount: big.NewInt(1e18)}
	defaultCmnCoins    = []cmn.Coin{baseDenomCmnCoin}
	atomCoins          = sdk.Coins{sdk.Coin{Denom: "uatom", Amount: math.NewInt(1e18)}}
	atomCmnCoin        = cmn.Coin{Denom: "uatom", Amount: big.NewInt(1e18)}
	mutliSpendLimit    = sdk.Coins{sdk.Coin{Denom: utils.BaseDenom, Amount: math.NewInt(1e18)}, sdk.Coin{Denom: "uatom", Amount: math.NewInt(1e18)}}
	mutliCmnCoins      = []cmn.Coin{baseDenomCmnCoin, atomCmnCoin}
	testERC20          = erc20Meta{
		Name:     "TestCoin",
		Symbol:   "TC",
		Decimals: 18,
	}
)

// setupIBCCoordinator sets up the IBC coordinator
func (s *PrecompileTestSuite) setupIBCCoordinator() {
	ibcSender, ibcSenderPrivKey := s.keyring.GetAccAddr(0), s.keyring.GetPrivKey(0)
	ibcAcc, err := s.grpcHandler.GetAccount(ibcSender.String())
	s.Require().NoError(err)

	IBCCoordinator := coordinator.NewIntegrationCoordinator(
		s.T(),
		[]commonnetwork.Network{s.network},
	)

	IBCCoordinator.SetDefaultSignerForChain(s.network.GetChainID(), ibcSenderPrivKey, ibcAcc)
	fmt.Println(s.network.GetChainID(), IBCCoordinator.GetDummyChainsIds()[0])
	IBCCoordinator.Setup(s.network.GetChainID(), IBCCoordinator.GetDummyChainsIds()[0])

	err = IBCCoordinator.CommitAll()
	s.Require().NoError(err)
}

// NewTransferAuthorizationWithAllocations creates a new allocation for the given grantee and granter and the given coins
func (s *PrecompileTestSuite) NewTransferAuthorizationWithAllocations(ctx sdk.Context, app *evmosapp.Evmos, grantee, granter common.Address, allocations []transfertypes.Allocation) error {
	transferAuthz := &transfertypes.TransferAuthorization{Allocations: allocations}
	if err := transferAuthz.ValidateBasic(); err != nil {
		return err
	}

	// create the authorization
	return app.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), transferAuthz, &s.defaultExpirationDuration)
}

// NewTransferAuthorization creates a new transfer authorization for the given grantee and granter and the given coins
func (s *PrecompileTestSuite) NewTransferAuthorization(ctx sdk.Context, app *evmosapp.Evmos, grantee, granter common.Address, path *ibctesting.Path, coins sdk.Coins, allowList []string) error {
	allocations := []transfertypes.Allocation{
		{
			SourcePort:    path.EndpointA.ChannelConfig.PortID,
			SourceChannel: path.EndpointA.ChannelID,
			SpendLimit:    coins,
			AllowList:     allowList,
		},
	}

	transferAuthz := &transfertypes.TransferAuthorization{Allocations: allocations}
	if err := transferAuthz.ValidateBasic(); err != nil {
		return err
	}

	// create the authorization
	return app.AuthzKeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), transferAuthz, &s.defaultExpirationDuration)
}

// GetTransferAuthorization returns the transfer authorization for the given grantee and granter
func (s *PrecompileTestSuite) GetTransferAuthorization(ctx sdk.Context, grantee, granter common.Address) *transfertypes.TransferAuthorization {
	grant, _ := s.network.App.AuthzKeeper.GetAuthorization(ctx, grantee.Bytes(), granter.Bytes(), ics20.TransferMsgURL)
	s.Require().NotNil(grant)
	transferAuthz, ok := grant.(*transfertypes.TransferAuthorization)
	s.Require().True(ok)
	s.Require().NotNil(transferAuthz)
	return transferAuthz
}

// CheckAllowanceChangeEvent is a helper function used to check the allowance change event arguments.
func (s *PrecompileTestSuite) CheckAllowanceChangeEvent(ctx sdk.Context, address common.Address, log *ethtypes.Log, amount *big.Int, isIncrease bool) {
	// Check event signature matches the one emitted
	event := s.precompile.ABI.Events[authorization.EventTypeIBCTransferAuthorization]
	s.Require().Equal(event.ID, common.HexToHash(log.Topics[0].Hex()))
	s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight()))

	var approvalEvent ics20.EventTransferAuthorization
	err := cmn.UnpackLog(s.precompile.ABI, &approvalEvent, authorization.EventTypeIBCTransferAuthorization, *log)
	s.Require().NoError(err)
	s.Require().Equal(address, approvalEvent.Grantee)
	s.Require().Equal(address, approvalEvent.Granter)
	s.Require().Equal("transfer", approvalEvent.Allocations[0].SourcePort)
	s.Require().Equal("channel-0", approvalEvent.Allocations[0].SourceChannel)

	allocationAmount := approvalEvent.Allocations[0].SpendLimit[0].Amount
	if isIncrease {
		newTotal := amount.Add(allocationAmount, amount)
		s.Require().Equal(amount, newTotal)
	} else {
		newTotal := amount.Sub(allocationAmount, amount)
		s.Require().Equal(amount, newTotal)
	}
}

// NewTransferPath creates a new path between two chains with the specified portIds and version.
func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = transfertypes.PortID
	path.EndpointB.ChannelConfig.PortID = transfertypes.PortID
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	return path
}

//// setupIBCTest makes the necessary setup of chains A & B
//// for integration tests
//func (s *PrecompileTestSuite) setupIBCTest() {
//	s.coordinator.CommitNBlocks(s.chainA, 2)
//	s.coordinator.CommitNBlocks(s.chainB, 2)
//
//	s.app = s.chainA.App.(*evmosapp.Evmos)
//	evmParams := s.app.EvmKeeper.GetParams(s.chainA.GetContext())
//	evmParams.EvmDenom = utils.BaseDenom
//	err := s.app.EvmKeeper.SetParams(s.chainA.GetContext(), evmParams)
//	s.Require().NoError(err)
//
//	// Set block proposer once, so its carried over on the ibc-go-testing suite
//	validators, err := s.app.StakingKeeper.GetValidators(s.chainA.GetContext(), 2)
//	s.Require().NoError(err)
//	cons, err := validators[0].GetConsAddr()
//	s.Require().NoError(err)
//	s.chainA.CurrentHeader.ProposerAddress = cons
//
//	err = s.app.StakingKeeper.SetValidatorByConsAddr(s.chainA.GetContext(), validators[0])
//	s.Require().NoError(err)
//
//	_, err = s.app.EvmKeeper.GetCoinbaseAddress(s.chainA.GetContext(), sdk.ConsAddress(s.chainA.CurrentHeader.ProposerAddress))
//	s.Require().NoError(err)
//
//	// Mint coins locked on the evmos account generated with secp.
//	amt, ok := math.NewIntFromString("1000000000000000000000")
//	s.Require().True(ok)
//	coinEvmos := sdk.NewCoin(utils.BaseDenom, amt)
//	coins := sdk.NewCoins(coinEvmos)
//	err = s.app.BankKeeper.MintCoins(s.chainA.GetContext(), inflationtypes.ModuleName, coins)
//	s.Require().NoError(err)
//	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.chainA.GetContext(), inflationtypes.ModuleName, s.chainA.SenderAccount.GetAddress(), coins)
//	s.Require().NoError(err)
//
//	s.transferPath = evmosibc.NewTransferPath(s.chainA, s.chainB) // clientID, connectionID, channelID empty
//	evmosibc.SetupPath(s.coordinator, s.transferPath)             // clientID, connectionID, channelID filled
//	s.Require().Equal("07-tendermint-0", s.transferPath.EndpointA.ClientID)
//	s.Require().Equal("connection-0", s.transferPath.EndpointA.ConnectionID)
//	s.Require().Equal("channel-0", s.transferPath.EndpointA.ChannelID)
//}

// setTransferApproval sets the transfer approval for the given grantee and allocations
//func (s *PrecompileTestSuite) setTransferApproval(
//	args contracts.CallArgs,
//	grantee common.Address,
//	allocations []cmn.ICS20Allocation,
//) {
//	args.MethodName = authorization.ApproveMethod
//	args.Args = []interface{}{
//		grantee,
//		allocations,
//	}
//
//	logCheckArgs := testutil.LogCheckArgs{
//		ABIEvents: s.precompile.Events,
//		ExpEvents: []string{authorization.EventTypeIBCTransferAuthorization},
//		ExpPass:   true,
//	}
//
//	_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, args, logCheckArgs)
//	Expect(err).To(BeNil(), "error while calling the contract to approve")
//
//	s.chainA.NextBlock()
//
//	// check auth created successfully
//	authz, _ := s.app.AuthzKeeper.GetAuthorization(s.chainA.GetContext(), grantee.Bytes(), args.PrivKey.PubKey().Address().Bytes(), ics20.TransferMsgURL)
//	Expect(authz).NotTo(BeNil())
//	transferAuthz, ok := authz.(*transfertypes.TransferAuthorization)
//	Expect(ok).To(BeTrue())
//	Expect(len(transferAuthz.Allocations[0].SpendLimit)).To(Equal(len(allocations[0].SpendLimit)))
//	for i, sl := range transferAuthz.Allocations[0].SpendLimit {
//		// NOTE order may change if there're more than one coin
//		Expect(sl.Denom).To(Equal(allocations[0].SpendLimit[i].Denom))
//		Expect(sl.Amount.BigInt()).To(Equal(allocations[0].SpendLimit[i].Amount))
//	}
//}

// setTransferApprovalForContract sets the transfer approval for the given contract
//func (s *PrecompileTestSuite) setTransferApprovalForContract(args contracts.CallArgs) {
//	logCheckArgs := testutil.LogCheckArgs{
//		ABIEvents: s.precompile.Events,
//		ExpEvents: []string{authorization.EventTypeIBCTransferAuthorization},
//		ExpPass:   true,
//	}
//
//	_, _, err := contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, args, logCheckArgs)
//	Expect(err).To(BeNil(), "error while calling the contract to approve")
//
//	s.chainA.NextBlock()
//
//	// check auth created successfully
//	authz, _ := s.app.AuthzKeeper.GetAuthorization(s.chainA.GetContext(), args.ContractAddr.Bytes(), args.PrivKey.PubKey().Address().Bytes(), ics20.TransferMsgURL)
//	Expect(authz).NotTo(BeNil())
//	transferAuthz, ok := authz.(*transfertypes.TransferAuthorization)
//	Expect(ok).To(BeTrue())
//	Expect(len(transferAuthz.Allocations) > 0).To(BeTrue())
//}

// setupAllocationsForTesting sets the allocations for testing
//func (s *PrecompileTestSuite) setupAllocationsForTesting() {
//	defaultSingleAlloc = []cmn.ICS20Allocation{
//		{
//			SourcePort:    ibctesting.TransferPort,
//			SourceChannel: s.transferPath.EndpointA.ChannelID,
//			SpendLimit:    defaultCmnCoins,
//		},
//	}
//}

// TODO upstream this change to evmos (adding gasPrice)
// DeployContract deploys a contract with the provided private key,
// compiled contract data and constructor arguments
func DeployContract(
	ctx sdk.Context,
	evmosApp *evmosapp.Evmos,
	priv cryptotypes.PrivKey,
	gasPrice *big.Int,
	queryClientEvm evmtypes.QueryClient,
	contract evmtypes.CompiledContract,
	constructorArgs ...interface{},
) (common.Address, error) {
	chainID := evmosApp.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := evmosApp.EvmKeeper.GetNonce(ctx, from)

	ctorArgs, err := contract.ABI.Pack("", constructorArgs...)
	if err != nil {
		return common.Address{}, err
	}

	data := append(contract.Bin, ctorArgs...) //nolint:gocritic
	gas, err := evmosutiltx.GasLimit(ctx, from, data, queryClientEvm)
	if err != nil {
		return common.Address{}, err
	}

	msgEthereumTx := evmtypes.NewTx(&evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		GasLimit:  gas,
		GasFeeCap: evmosApp.FeeMarketKeeper.GetBaseFee(ctx),
		GasTipCap: big.NewInt(1),
		GasPrice:  gasPrice,
		Input:     data,
		Accesses:  &ethtypes.AccessList{},
	})
	msgEthereumTx.From = from.String()

	res, err := evmosutil.DeliverEthTx(evmosApp, priv, msgEthereumTx)
	if err != nil {
		return common.Address{}, err
	}

	if _, err := evmosutil.CheckEthTxResponse(res, evmosApp.AppCodec()); err != nil {
		return common.Address{}, err
	}

	return crypto.CreateAddress(from, nonce), nil
}

//// DeployERC20Contract deploys a ERC20 token with the provided name, symbol and decimals
//func (s *PrecompileTestSuite) DeployERC20Contract(chain *ibctesting.TestChain, name, symbol string, decimals uint8) (common.Address, error) {
//	addr, err := DeployContract(
//		chain.GetContext(),
//		s.app,
//		s.privKey,
//		gasPrice,
//		s.queryClientEVM,
//		evmoscontracts.ERC20MinterBurnerDecimalsContract,
//		name,
//		symbol,
//		decimals,
//	)
//	chain.NextBlock()
//	return addr, err
//}
//
//// setupERC20ContractTests deploys a ERC20 token
//// and mint some tokens to the deployer address (s.address).
//// The amount of tokens sent to the deployer address is defined in
//// the 'amount' input argument
//func (s *PrecompileTestSuite) setupERC20ContractTests(amount *big.Int) common.Address {
//	erc20Addr, err := s.DeployERC20Contract(s.chainA, testERC20.Name, testERC20.Symbol, testERC20.Decimals)
//	Expect(err).To(BeNil(), "error while deploying ERC20 contract: %v", err)
//
//	defaultERC20CallArgs := contracts.CallArgs{
//		ContractAddr: erc20Addr,
//		ContractABI:  evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
//		PrivKey:      s.privKey,
//		GasPrice:     gasPrice,
//	}
//
//	// mint coins to the address
//	mintCoinsArgs := defaultERC20CallArgs.
//		WithMethodName("mint").
//		WithArgs(s.address, amount)
//
//	mintCheck := testutil.LogCheckArgs{
//		ABIEvents: evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI.Events,
//		ExpEvents: []string{erc20.EventTypeTransfer}, // upon minting the tokens are sent to the receiving address
//		ExpPass:   true,
//	}
//
//	_, _, err = contracts.CallContractAndCheckLogs(s.chainA.GetContext(), s.app, mintCoinsArgs, mintCheck)
//	Expect(err).To(BeNil(), "error while calling the smart contract: %v", err)
//
//	s.chainA.NextBlock()
//
//	// check that the address has the tokens -- this has to be done using the stateDB because
//	// unregistered token pairs do not show up in the bank keeper
//	balance := s.app.Erc20Keeper.BalanceOf(
//		s.chainA.GetContext(),
//		evmoscontracts.ERC20MinterBurnerDecimalsContract.ABI,
//		erc20Addr,
//		s.address,
//	)
//	Expect(balance).To(Equal(amount), "address does not have the expected amount of tokens")
//
//	return erc20Addr
//}
//
//// makePacket is a helper function to build the sent IBC packet
//// to perform an ICS20 transfer.
//// This packet is then used to test the IBC callbacks (Timeout, Ack)
//func (s *PrecompileTestSuite) makePacket(
//	senderAddr,
//	receiverAddr,
//	denom,
//	memo string,
//	amt *big.Int,
//	seq uint64,
//	timeoutHeight clienttypes.Height,
//) channeltypes.Packet {
//	packetData := transfertypes.NewFungibleTokenPacketData(
//		denom,
//		amt.String(),
//		senderAddr,
//		receiverAddr,
//		memo,
//	)
//
//	return channeltypes.NewPacket(
//		packetData.GetBytes(),
//		seq,
//		s.transferPath.EndpointA.ChannelConfig.PortID,
//		s.transferPath.EndpointA.ChannelID,
//		s.transferPath.EndpointB.ChannelConfig.PortID,
//		s.transferPath.EndpointB.ChannelID,
//		timeoutHeight,
//		0,
//	)
//}
