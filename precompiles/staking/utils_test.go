package staking_test

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"slices"
	"time"

	cmnfactory "github.com/evmos/evmos/v20/testutil/integration/common/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/factory"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/grpc"
	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"

	//nolint:revive // dot imports are fine for Ginkgo
	. "github.com/onsi/gomega"

	"cosmossdk.io/math"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/evmos/evmos/v20/precompiles/authorization"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/precompiles/staking"
	"github.com/evmos/evmos/v20/precompiles/testutil"
	evmosutil "github.com/evmos/evmos/v20/testutil"
	"github.com/evmos/evmos/v20/utils"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	stakingkeeper "github.com/evmos/evmos/v20/x/staking/keeper"
	vestingtypes "github.com/evmos/evmos/v20/x/vesting/types"
)

// stipend to pay EVM tx fees
var (
	accountGasCoverage = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e16)))
	gas                = uint64(200_000)
	gasPrices          = accountGasCoverage.QuoInt(math.NewIntFromUint64(gas)).AmountOf(utils.BaseDenom)
)

// ApproveAndCheckAuthz is a helper function to approve a given authorization method and check if the authorization was created.
func (s *PrecompileTestSuite) ApproveAndCheckAuthz(method abi.Method, granter, grantee testkeyring.Key, msgType string, amount *big.Int) {
	approveArgs := []interface{}{
		grantee.Addr,
		amount,
		[]string{msgType},
	}
	resp, err := s.precompile.Approve(s.network.GetContext(), granter.Addr, s.network.GetStateDB(), &method, approveArgs)
	s.Require().NoError(err)
	s.Require().Equal(resp, cmn.TrueValue)

	auth, _ := CheckAuthorizationWithContext(s.network.GetContext(), s.network.App.AuthzKeeper, staking.DelegateAuthz, grantee.Addr, granter.Addr)
	s.Require().NotNil(auth)
	s.Require().Equal(auth.AuthorizationType, staking.DelegateAuthz)
	s.Require().Equal(auth.MaxTokens, &sdk.Coin{Denom: s.bondDenom, Amount: math.NewIntFromBigInt(amount)})
}

// CheckAuthorizationWithContext is a helper function to check if the authorization is set and if it is the correct type.
// Useful only for unit tests
func CheckAuthorizationWithContext(ctx sdk.Context, ak authzkeeper.Keeper, authorizationType stakingtypes.AuthorizationType, grantee, granter common.Address) (*stakingtypes.StakeAuthorization, *time.Time) {
	stakingAuthz := stakingtypes.StakeAuthorization{AuthorizationType: authorizationType}
	auth, expirationTime := ak.GetAuthorization(ctx, grantee.Bytes(), granter.Bytes(), stakingAuthz.MsgTypeURL())

	stakeAuthorization, ok := auth.(*stakingtypes.StakeAuthorization)
	if !ok {
		return nil, expirationTime
	}

	return stakeAuthorization, expirationTime
}

// CheckAuthorization is a helper function to check if the authorization is set and if it is the correct type.
func CheckAuthorization(gh grpc.Handler, ir cdctypes.InterfaceRegistry, authorizationType stakingtypes.AuthorizationType, grantee, granter common.Address) (*stakingtypes.StakeAuthorization, *time.Time, error) {
	grants, err := gh.GetGrants(sdk.AccAddress(grantee.Bytes()).String(), sdk.AccAddress(granter.Bytes()).String())
	if err != nil {
		return nil, nil, err
	}

	if len(grants) == 0 {
		return nil, nil, fmt.Errorf("no authorizations found for grantee %s and granter %s", grantee, granter)
	}

	var (
		expGrant           *authz.Grant
		stakeAuthorization *stakingtypes.StakeAuthorization
	)
	for _, g := range grants {
		var (
			ok   bool
			auth authz.Authorization
		)
		if err = ir.UnpackAny(g.Authorization, &auth); err != nil {
			return nil, nil, err
		}
		stakeAuthorization, ok = auth.(*stakingtypes.StakeAuthorization)
		if !ok {
			return nil, nil, fmt.Errorf("invalid authorization type. Expected: stakingtypes.StakeAuthorization, got: %T", auth)
		}
		if stakeAuthorization.AuthorizationType == authorizationType {
			expGrant = g
			break
		}
	}

	if expGrant == nil {
		return nil, nil, fmt.Errorf("invalid authorization type. Expected: %d, got: %d", authorizationType, stakeAuthorization.AuthorizationType)
	}

	return stakeAuthorization, expGrant.Expiration, nil
}

// CreateAuthorization is a helper function to create a new authorization of the given type for a spender address
// (=grantee).
// The authorization will be created to spend the given Coin.
// For testing purposes, this function will create a new authorization for all available validators,
// that are not jailed.
func (s *PrecompileTestSuite) CreateAuthorization(ctx sdk.Context, granter, grantee sdk.AccAddress, authzType stakingtypes.AuthorizationType, coin *sdk.Coin) error {
	// Get all available validators and filter out jailed validators
	validators := make([]sdk.ValAddress, 0)
	err := s.network.App.StakingKeeper.IterateValidators(
		ctx, func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {
			if validator.IsJailed() {
				return
			}
			validators = append(validators, sdk.ValAddress(validator.GetOperator()))
			return
		},
	)
	if err != nil {
		return err
	}

	stakingAuthz, err := stakingtypes.NewStakeAuthorization(validators, nil, authzType, coin)
	if err != nil {
		return err
	}

	expiration := time.Now().Add(cmn.DefaultExpirationDuration).UTC()
	err = s.network.App.AuthzKeeper.SaveGrant(ctx, grantee, granter, stakingAuthz, &expiration)
	if err != nil {
		return err
	}

	return nil
}

// SetupApproval sets up an approval, that authorizes the grantee to spend the given amount for the granter
// in transactions, that target the specified message types.
func (s *PrecompileTestSuite) SetupApproval(
	granterPriv types.PrivKey,
	grantee common.Address,
	amount *big.Int,
	msgTypes []string,
) {
	precompileAddr := s.precompile.Address()
	txArgs := evmtypes.EvmTxArgs{
		To: &precompileAddr,
	}
	approveArgs := factory.CallArgs{
		ContractABI: s.precompile.ABI,
		MethodName:  authorization.ApproveMethod,
		Args: []interface{}{
			grantee, amount, msgTypes,
		},
	}

	logCheckArgs := testutil.LogCheckArgs{
		ABIEvents: s.precompile.Events,
		ExpEvents: []string{authorization.EventTypeApproval},
		ExpPass:   true,
	}

	res, _, err := s.factory.CallContractAndCheckLogs(
		granterPriv,
		txArgs, approveArgs,
		logCheckArgs,
	)
	Expect(err).To(BeNil(), "error while calling the contract to approve")
	Expect(s.network.NextBlock()).To(BeNil())

	// Check if the approval event is emitted
	granterAddr := common.BytesToAddress(granterPriv.PubKey().Address().Bytes())
	testutil.CheckAuthorizationEvents(
		s.precompile.Events[authorization.EventTypeApproval],
		s.precompile.Address(),
		granterAddr,
		grantee,
		res,
		s.network.GetContext().BlockHeight(),
		msgTypes,
		amount,
	)
}

// SetupApprovalWithContractCalls is a helper function used to setup the allowance for the given spender.
func (s *PrecompileTestSuite) SetupApprovalWithContractCalls(
	granter testkeyring.Key,
	txArgs evmtypes.EvmTxArgs,
	approvalArgs factory.CallArgs,
) {
	msgTypes, ok := approvalArgs.Args[1].([]string)
	Expect(ok).To(BeTrue(), "failed to convert msgTypes to []string")
	expAmount, ok := approvalArgs.Args[2].(*big.Int)
	Expect(ok).To(BeTrue(), "failed to convert amount to big.Int")

	logCheckArgs := testutil.LogCheckArgs{
		ABIEvents: s.precompile.Events,
		ExpEvents: []string{authorization.EventTypeApproval},
		ExpPass:   true,
	}

	_, _, err := s.factory.CallContractAndCheckLogs(
		granter.Priv,
		txArgs,
		approvalArgs,
		logCheckArgs,
	)
	Expect(err).To(BeNil(), "error while approving: %v", err)
	Expect(s.network.NextBlock()).To(BeNil())

	// iterate over args
	var expectedAuthz stakingtypes.AuthorizationType
	for _, msgType := range msgTypes {
		switch msgType {
		case staking.DelegateMsg:
			expectedAuthz = staking.DelegateAuthz
		case staking.UndelegateMsg:
			expectedAuthz = staking.UndelegateAuthz
		case staking.RedelegateMsg:
			expectedAuthz = staking.RedelegateAuthz
		case staking.CancelUnbondingDelegationMsg:
			expectedAuthz = staking.CancelUnbondingDelegationAuthz
		}
		authz, expirationTime, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, expectedAuthz, *txArgs.To, granter.Addr)
		Expect(err).To(BeNil())
		Expect(authz).ToNot(BeNil(), "expected authorization to be set")
		Expect(authz.MaxTokens.Amount).To(Equal(math.NewInt(expAmount.Int64())), "expected different allowance")
		Expect(authz.MsgTypeURL()).To(Equal(msgType), "expected different message type")
		Expect(expirationTime).ToNot(BeNil(), "expected expiration time to not be nil")
	}
}

// CheckAllowanceChangeEvent is a helper function used to check the allowance change event arguments.
func (s *PrecompileTestSuite) CheckAllowanceChangeEvent(
	log *ethtypes.Log, methods []string, amounts []*big.Int, granter, grantee common.Address,
) {
	s.Require().Equal(log.Address, s.precompile.Address())
	// Check event signature matches the one emitted
	event := s.precompile.ABI.Events[authorization.EventTypeAllowanceChange]
	s.Require().Equal(event.ID, common.HexToHash(log.Topics[0].Hex()))
	s.Require().Equal(log.BlockNumber, uint64(s.network.GetContext().BlockHeight())) //nolint:gosec // G115

	var approvalEvent authorization.EventAllowanceChange
	err := cmn.UnpackLog(s.precompile.ABI, &approvalEvent, authorization.EventTypeAllowanceChange, *log)
	s.Require().NoError(err)
	s.Require().Equal(grantee, approvalEvent.Grantee)
	s.Require().Equal(granter, approvalEvent.Granter)
	s.Require().Equal(len(methods), len(approvalEvent.Methods))

	for i, method := range methods {
		s.Require().Equal(method, approvalEvent.Methods[i])
		s.Require().Equal(amounts[i], approvalEvent.Values[i])
	}
}

// ExpectAuthorization is a helper function for tests using the Ginkgo BDD style tests, to check that the
// authorization is correctly set.
func (s *PrecompileTestSuite) ExpectAuthorization(authorizationType stakingtypes.AuthorizationType, grantee, granter common.Address, maxTokens *sdk.Coin) {
	authz, expirationTime, err := CheckAuthorization(s.grpcHandler, s.network.GetEncodingConfig().InterfaceRegistry, authorizationType, grantee, granter)
	Expect(err).To(BeNil())
	Expect(authz).ToNot(BeNil(), "expected authorization to be set")
	Expect(authz.AuthorizationType).To(Equal(authorizationType), "expected different authorization type")
	Expect(authz.MaxTokens).To(Equal(maxTokens), "expected different max tokens")
	Expect(expirationTime).ToNot(BeNil(), "expected expiration time to be not be nil")
}

// assertValidatorsResponse asserts all the fields on the validators response
func (s *PrecompileTestSuite) assertValidatorsResponse(validators []staking.ValidatorInfo, expLen int) {
	// returning order can change
	valOrder := []int{0, 1}
	varAddr := sdk.ValAddress(common.HexToAddress(validators[0].OperatorAddress).Bytes()).String()
	vals := s.network.GetValidators()

	if varAddr != vals[0].OperatorAddress {
		valOrder = []int{1, 0}
	}
	for i := 0; i < expLen; i++ {
		j := valOrder[i]

		val := s.network.GetValidators()[j]
		s.Require().Equal(val.OperatorAddress, sdk.ValAddress(common.HexToAddress(validators[i].OperatorAddress).Bytes()).String())
		s.Require().Equal(uint8(val.Status), validators[i].Status) //#nosec G115
		s.Require().Equal(val.Tokens.Uint64(), validators[i].Tokens.Uint64())
		s.Require().Equal(val.DelegatorShares.BigInt(), validators[i].DelegatorShares)
		s.Require().Equal(val.Jailed, validators[i].Jailed)
		s.Require().Equal(val.UnbondingHeight, validators[i].UnbondingHeight)
		s.Require().Equal(int64(0), validators[i].UnbondingTime)
		s.Require().Equal(math.LegacyNewDecWithPrec(5, 2).BigInt(), validators[i].Commission)
		s.Require().Equal(int64(0), validators[i].MinSelfDelegation.Int64())
		s.Require().Equal(validators[i].ConsensusPubkey, staking.FormatConsensusPubkey(val.ConsensusPubkey))
	}
}

// assertRedelegation asserts the redelegationOutput struct and its fields
func (s *PrecompileTestSuite) assertRedelegationsOutput(data []byte, redelTotalCount uint64, expAmt *big.Int, expCreationHeight int64, hasPagination bool) {
	var redOut staking.RedelegationsOutput
	err := s.precompile.UnpackIntoInterface(&redOut, staking.RedelegationsMethod, data)
	s.Require().NoError(err, "failed to unpack output")

	s.Require().Len(redOut.Response, 1)
	// check pagination - total count should be 2
	s.Require().Equal(redelTotalCount, redOut.PageResponse.Total)
	if hasPagination {
		s.Require().NotEmpty(redOut.PageResponse.NextKey)
	} else {
		s.Require().Empty(redOut.PageResponse.NextKey)
	}
	// check redelegation entry
	// order may change, one redelegation has 2 entries
	// and the other has one
	if len(redOut.Response[0].Entries) == 2 {
		s.assertRedelegation(redOut.Response[0],
			2,
			s.network.GetValidators()[0].OperatorAddress,
			s.network.GetValidators()[1].OperatorAddress,
			expAmt,
			expCreationHeight,
		)
	} else {
		s.assertRedelegation(redOut.Response[0],
			1,
			s.network.GetValidators()[0].OperatorAddress,
			s.network.GetValidators()[2].OperatorAddress,
			expAmt,
			expCreationHeight,
		)
	}
}

// assertRedelegation asserts all the fields on the redelegations response
// should specify the amount of entries expected and the expected amount for this
// the same amount is considered for all entries
func (s *PrecompileTestSuite) assertRedelegation(res staking.RedelegationResponse, entriesCount int, expValSrcAddr, expValDstAddr string, expAmt *big.Int, expCreationHeight int64) {
	// check response
	s.Require().Equal(res.Redelegation.DelegatorAddress, s.keyring.GetAccAddr(0).String())
	s.Require().Equal(res.Redelegation.ValidatorSrcAddress, expValSrcAddr)
	s.Require().Equal(res.Redelegation.ValidatorDstAddress, expValDstAddr)
	// check redelegation entries - should be empty
	s.Require().Empty(res.Redelegation.Entries)
	// check response entries, should be 2
	s.Require().Len(res.Entries, entriesCount)
	// check redelegation entries
	for _, e := range res.Entries {
		s.Require().Equal(e.Balance, expAmt)
		s.Require().True(e.RedelegationEntry.CompletionTime > 1600000000)
		s.Require().Equal(expCreationHeight, e.RedelegationEntry.CreationHeight)
		s.Require().Equal(e.RedelegationEntry.InitialBalance, expAmt)
	}
}

// setupRedelegations setups 2 entries for redelegation from validator[0]
// to validator[1], and a redelegation from validator[0] to validator[2]
func (s *PrecompileTestSuite) setupRedelegations(ctx sdk.Context, redelAmt *big.Int) error {
	ctx = ctx.WithBlockTime(time.Now())
	vals := s.network.GetValidators()

	msg := stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    s.keyring.GetAccAddr(0).String(),
		ValidatorSrcAddress: vals[0].OperatorAddress,
		ValidatorDstAddress: vals[1].OperatorAddress,
		Amount:              sdk.NewCoin(s.bondDenom, math.NewIntFromBigInt(redelAmt)),
	}

	msgSrv := stakingkeeper.NewMsgServerImpl(&s.network.App.StakingKeeper)
	// create 2 entries for same redelegation
	for i := 0; i < 2; i++ {
		if _, err := msgSrv.BeginRedelegate(ctx, &msg); err != nil {
			return err
		}
	}

	// create a redelegation from validator[0] to validator[2]
	msg.ValidatorDstAddress = vals[2].OperatorAddress
	_, err := msgSrv.BeginRedelegate(ctx, &msg)
	return err
}

// CheckValidatorOutput checks that the given validator output
func (s *PrecompileTestSuite) CheckValidatorOutput(valOut staking.ValidatorInfo) {
	vals := s.network.GetValidators()
	validatorAddrs := make([]string, len(vals))
	for i, v := range vals {
		validatorAddrs[i] = v.OperatorAddress
	}

	operatorAddress := sdk.ValAddress(common.HexToAddress(valOut.OperatorAddress).Bytes()).String()

	Expect(slices.Contains(validatorAddrs, operatorAddress)).To(BeTrue(), "operator address not found in test suite validators")
	Expect(valOut.DelegatorShares).To(Equal(big.NewInt(1e18)), "expected different delegator shares")
}

// setupVestingAccount is a helper function used in integraiton tests to setup a vesting account
// using the TestVestingSchedule. Also, funds the account with extra funds to pay for transaction fees
func (s *PrecompileTestSuite) setupVestingAccount(funder, vestAcc testkeyring.Key) *vestingtypes.ClawbackVestingAccount {
	vestingAmtTotal := evmosutil.TestVestingSchedule.TotalVestingCoins
	ctx := s.network.GetContext()

	// send some funds to the vesting acccount to pay for fees
	err := s.factory.FundAccount(funder, vestAcc.AccAddr, accountGasCoverage)
	Expect(err).To(BeNil())
	Expect(s.network.NextBlock()).To(BeNil())

	// 1. Create vesting account
	createAccMsg := vestingtypes.NewMsgCreateClawbackVestingAccount(
		funder.AccAddr,
		vestAcc.AccAddr,
		false,
	)

	_, err = s.factory.ExecuteCosmosTx(vestAcc.Priv, cmnfactory.CosmosTxArgs{Msgs: []sdk.Msg{createAccMsg}, Gas: &gas, GasPrice: &gasPrices})
	Expect(err).To(BeNil())
	Expect(s.network.NextBlock()).To(BeNil())

	// 2. Funder funds the vesting account
	vestingStart := ctx.BlockTime()
	fundMsg := vestingtypes.NewMsgFundVestingAccount(
		funder.AccAddr,
		vestAcc.AccAddr,
		vestingStart,
		evmosutil.TestVestingSchedule.LockupPeriods,
		evmosutil.TestVestingSchedule.VestingPeriods,
	)
	_, err = s.factory.ExecuteCosmosTx(funder.Priv, cmnfactory.CosmosTxArgs{Msgs: []sdk.Msg{fundMsg}})
	Expect(err).To(BeNil())
	Expect(s.network.NextBlock()).To(BeNil())

	acc, err := s.grpcHandler.GetAccount(vestAcc.AccAddr.String())
	Expect(err).To(BeNil())

	clawbackAccount, ok := acc.(*vestingtypes.ClawbackVestingAccount)
	Expect(ok).To(BeTrue(), "account should be a ClawbackVestingAccount")

	// Check all coins are locked up
	lockedUp := clawbackAccount.GetLockedUpCoins(ctx.BlockTime())
	Expect(vestingAmtTotal).To(Equal(lockedUp))

	// Grant gas stipend to cover EVM fees
	err = s.factory.FundAccount(funder, vestAcc.AccAddr, accountGasCoverage)
	Expect(err).To(BeNil())
	Expect(s.network.NextBlock()).To(BeNil())

	balRes, err := s.grpcHandler.GetBalance(clawbackAccount.GetAddress(), s.bondDenom)
	Expect(err).To(BeNil())
	Expect(*balRes.Balance).To(Equal(accountGasCoverage[0].Add(vestingAmtTotal[0])))

	return clawbackAccount
}

// Generate the Base64 encoded PubKey associated with a PrivKey generated with
// the ed25519 algorithm used in Tendermint nodes.
func GenerateBase64PubKey() string {
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey().(*ed25519.PubKey)
	return base64.StdEncoding.EncodeToString(pubKey.Bytes())
}
