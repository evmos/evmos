// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package grpc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v20/testutil/integration/common/network"
)

// Handler is an interface that defines the common methods that are used to query
// the network's modules via gRPC.
type Handler interface {
	// Account methods
	GetAccount(address string) (sdk.AccountI, error)

	// Authz methods
	GetAuthorizations(grantee, granter string) ([]authz.Authorization, error)
	GetAuthorizationsByGrantee(grantee string) ([]authz.Authorization, error)
	GetAuthorizationsByGranter(granter string) ([]authz.Authorization, error)
	GetGrants(grantee, granter string) ([]*authz.Grant, error)
	GetGrantsByGrantee(grantee string) ([]*authz.GrantAuthorization, error)
	GetGrantsByGranter(granter string) ([]*authz.GrantAuthorization, error)

	// Bank methods
	GetBalance(address sdk.AccAddress, denom string) (*banktypes.QueryBalanceResponse, error)
	GetSpendableBalance(address sdk.AccAddress, denom string) (*banktypes.QuerySpendableBalanceByDenomResponse, error)
	GetAllBalances(address sdk.AccAddress) (*banktypes.QueryAllBalancesResponse, error)
	GetTotalSupply() (*banktypes.QueryTotalSupplyResponse, error)

	// Staking methods
	GetDelegation(delegatorAddress string, validatorAddress string) (*stakingtypes.QueryDelegationResponse, error)
	GetDelegatorDelegations(delegatorAddress string) (*stakingtypes.QueryDelegatorDelegationsResponse, error)
	GetValidatorDelegations(validatorAddress string) (*stakingtypes.QueryValidatorDelegationsResponse, error)
	GetRedelegations(delegatorAddress, srcValidator, dstValidator string) (*stakingtypes.QueryRedelegationsResponse, error)
	GetValidatorUnbondingDelegations(validatorAddress string) (*stakingtypes.QueryValidatorUnbondingDelegationsResponse, error)
	GetDelegatorUnbondingDelegations(delegatorAddress string) (*stakingtypes.QueryDelegatorUnbondingDelegationsResponse, error)

	// Distribution methods
	GetDelegationTotalRewards(delegatorAddress string) (*distrtypes.QueryDelegationTotalRewardsResponse, error)
	GetDelegationRewards(delegatorAddress string, validatorAddress string) (*distrtypes.QueryDelegationRewardsResponse, error)
	GetDelegatorWithdrawAddr(delegatorAddress string) (*distrtypes.QueryDelegatorWithdrawAddressResponse, error)
	GetValidatorCommission(validatorAddress string) (*distrtypes.QueryValidatorCommissionResponse, error)
	GetValidatorOutstandingRewards(validatorAddress string) (*distrtypes.QueryValidatorOutstandingRewardsResponse, error)
	GetCommunityPool() (*distrtypes.QueryCommunityPoolResponse, error)
	GetBondedValidators() (*stakingtypes.QueryValidatorsResponse, error)
}

var _ Handler = (*IntegrationHandler)(nil)

// IntegrationHandler is a helper struct to query the network's modules
// via gRPC. This is to simulate the behavior of a real user and avoid querying
// the modules directly.
type IntegrationHandler struct {
	network network.Network
}

// NewIntegrationHandler creates a new IntegrationHandler instance.
func NewIntegrationHandler(network network.Network) *IntegrationHandler {
	return &IntegrationHandler{
		network: network,
	}
}
