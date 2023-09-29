package v14_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmosapp "github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v14/testutil"
	testutiltx "github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/utils"
)

// CreateDelegationWithZeroTokens is a helper script, which creates a delegation
// so that the shares are not zero but the query for token amount returns zero tokens.
//
// NOTE: This is replicating an edge case that was found in mainnet data, which led to
// the account migrations not succeeding.
func CreateDelegationWithZeroTokens(
	ctx sdk.Context,
	app *evmosapp.Evmos,
	priv *ethsecp256k1.PrivKey,
	delegator sdk.AccAddress,
	validator stakingtypes.Validator,
) (stakingtypes.Delegation, error) {
	msgDelegate := stakingtypes.NewMsgDelegate(delegator, validator.GetOperator(), sdk.NewCoin(utils.BaseDenom, sdk.ZeroInt()))
	_, err := testutil.DeliverTx(ctx, app, priv, nil, msgDelegate)
	if err != nil {
		return stakingtypes.Delegation{}, fmt.Errorf("failed to delegate: %w", err)
	}

	delegation, found := app.StakingKeeper.GetDelegation(s.ctx, delegator, validator.GetOperator())
	if !found {
		return stakingtypes.Delegation{}, fmt.Errorf("delegation not found")
	}

	return delegation, nil
}

func (s *UpgradesTestSuite) TestCreateDelegationWithZeroTokens() {
	s.SetupTest()

	addr, priv := testutiltx.NewAccAddressAndKey()
	targetValidator := s.validators[1]

	delegation, err := CreateDelegationWithZeroTokens(s.ctx, s.app, priv, addr, targetValidator)
	s.Require().NoError(err, "failed to create delegation with zero tokens")
	s.Require().NotEqual(sdk.ZeroDec(), delegation.Shares, "delegation shares should not be zero")

	// Check that the validators tokenFromShares method returns zero tokens when truncated to an int
	tokens := targetValidator.TokensFromShares(delegation.Shares).TruncateInt()
	s.Require().Equal(int64(0), tokens.Int64(), "expected zero tokens to be returned")
}
