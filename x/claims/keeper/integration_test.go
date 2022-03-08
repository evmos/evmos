package keeper_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/tests"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/v2/app"
	"github.com/tharsis/evmos/v2/app/ante"
	"github.com/tharsis/evmos/v2/testutil"
	inflationtypes "github.com/tharsis/evmos/v2/x/inflation/types"

	"github.com/tharsis/evmos/v2/x/claims/types"
)

// TODO
// params := types.DefaultParams()
// params.EnableClaims = false
// s.app.ClaimsKeeper.SetParams(s.ctx, params)

var _ = Describe("Check amount claimed depending on claim time", Ordered, func() {
	params := types.DefaultParams()
	claimsAddr := s.app.AccountKeeper.GetModuleAddress(types.ModuleName)
	claimsAmount := int64(10000000)

	claimValue := int64(10000)
	actionValue := int64(claimValue / 4)

	addr1 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr3 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	addr4 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	delegationValue := sdk.NewInt(1)
	stakeDenom := stakingtypes.DefaultParams().BondDenom
	delegationAmount := sdk.NewCoins(sdk.NewCoin(stakeDenom, delegationValue))

	var (
		claimsRecord1 types.ClaimsRecord
		claimsRecord2 types.ClaimsRecord
		claimsRecord3 types.ClaimsRecord
		claimsRecord4 types.ClaimsRecord
	)

	BeforeEach(func() {
		s.SetupTest()

		params := s.app.ClaimsKeeper.GetParams(s.ctx)
		params.EnableClaims = true
		params.AirdropStartTime = s.ctx.BlockTime()
		s.app.ClaimsKeeper.SetParams(s.ctx, params)

		coins := sdk.NewCoins(sdk.NewCoin(params.GetClaimsDenom(), sdk.NewInt(claimsAmount)))

		err := s.app.BankKeeper.MintCoins(s.ctx, inflationtypes.ModuleName, coins)
		s.Require().NoError(err)
		err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, inflationtypes.ModuleName, types.ModuleName, coins)
		s.Require().NoError(err)

		balanceClaims := s.app.BankKeeper.GetBalance(s.ctx, claimsAddr, params.GetClaimsDenom())
		Expect(balanceClaims.Amount.Uint64()).To(Equal(uint64(claimsAmount)))

		testutil.FundAccount(s.app.BankKeeper, s.ctx, addr1, delegationAmount)

		claimsRecord1 = types.NewClaimsRecord(sdk.NewInt(claimValue))
		s.app.ClaimsKeeper.SetClaimsRecord(s.ctx, addr1, claimsRecord1)
		claimsRecord2 = types.NewClaimsRecord(sdk.NewInt(claimValue))
		s.app.ClaimsKeeper.SetClaimsRecord(s.ctx, addr2, claimsRecord2)
		claimsRecord3 = types.NewClaimsRecord(sdk.NewInt(claimValue))
		s.app.ClaimsKeeper.SetClaimsRecord(s.ctx, addr3, claimsRecord3)
		claimsRecord4 = types.NewClaimsRecord(sdk.NewInt(claimValue))
		s.app.ClaimsKeeper.SetClaimsRecord(s.ctx, addr4, claimsRecord4)

		balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())
		Expect(balance.Amount.Uint64()).To(Equal(uint64(0)))
		balance = s.app.BankKeeper.GetBalance(s.ctx, addr2, params.GetClaimsDenom())
		Expect(balance.Amount.Uint64()).To(Equal(uint64(0)))
		balance = s.app.BankKeeper.GetBalance(s.ctx, addr3, params.GetClaimsDenom())
		Expect(balance.Amount.Uint64()).To(Equal(uint64(0)))
		balance = s.app.BankKeeper.GetBalance(s.ctx, addr4, params.GetClaimsDenom())
		Expect(balance.Amount.Uint64()).To(Equal(uint64(0)))
	})

	Context("Claim amount claimed before decay duration  ", func() {

		It("Successfully claim action ActionDelegate", func() {
			delegate(addr1, 1)

			balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, params.GetClaimsDenom())
			Expect(balance.Amount.Uint64()).To(Equal(uint64(actionValue)))

			// _, err := s.app.ClaimsKeeper.ClaimCoinsForAction(s.ctx, addr1, claimsRecord1, types.ActionDelegate, params)
			// Expect(err).ToNot(BeNil())
		})
	})

	Context("Check amount claimed at 1/2 decay duration  ", func() {
		// cliffDuration := time.Duration(cliffLength)
		// s.CommitAfter(cliffDuration * time.Second)

	})

	Context("Check amount clawed back after decay duration  ", func() {
		// check community pool
		// balanceCommunityPool := s.app.DistrKeeper.GetFeePoolCommunityCoins(s.ctx)
	})
})

func nextFn(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

func delegate(accountAddress sdk.AccAddress, amount int64) error {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	addr, err := sdk.AccAddressFromBech32(accountAddress.String())
	s.Require().NoError(err)
	//
	val, err := sdk.ValAddressFromBech32("evmosvaloper1z3t55m0l9h0eupuz3dp5t5cypyv674jjn4d6nn")
	s.Require().NoError(err)
	delegateMsg := stakingtypes.NewMsgDelegate(addr, val, sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, sdk.NewInt(amount)))

	txBuilder.SetMsgs(delegateMsg)
	tx := txBuilder.GetTx()

	// s.app.StakingKeeper.AfterDelegationModified(s.ctx, accountAddress, s.app.validator)

	// Call Ante decorator
	dec := ante.NewEthVestingTransactionDecorator(s.app.AccountKeeper)
	_, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	return err

	// options := ante.HandlerOptions{
	// 	AccountKeeper:    s.app.AccountKeeper,
	// 	BankKeeper:       s.app.BankKeeper,
	// 	EvmKeeper:        s.app.EvmKeeper,
	// 	StakingKeeper:    s.app.StakingKeeper,
	// 	// FeegrantKeeper:   app.FeeGrantKeeper,
	// 	// IBCChannelKeeper: app.IBCKeeper.ChannelKeeper,
	// 	// FeeMarketKeeper:  app.FeeMarketKeeper,
	// 	// SignModeHandler:  encodingConfig.TxConfig.SignModeHandler(),
	// 	// SigGasConsumer:   SigVerificationGasConsumer,
	// 	// Cdc:              appCodec,
	// }

	// ante := ante.NewAnteHandler(options)
	// s.app.SetAnteHandler(ante)

	// sdktestutil.

	// _, err = ante.AnteHandle(s.ctx, tx, false, nextFn)
	// return err
}

func performEthTx(account *types.ClaimsRecordAddress) error {
	addr, err := sdk.AccAddressFromBech32(account.Address)
	s.Require().NoError(err)
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(addr.Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	msgEthereumTx := evmtypes.NewTx(chainID, nonce, &from, nil, 100000, nil, s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), nil, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msgEthereumTx)
	tx := txBuilder.GetTx()

	// Call Ante decorator
	dec := ante.NewEthVestingTransactionDecorator(s.app.AccountKeeper)
	_, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	return err

	// // Call Ante decorator
	// dec := ante.NewAnteHandler(s.app.AccountKeeper)
	// _, err = dec.AnteHandle(s.ctx, tx, false, nextFn)
	// return err
}
