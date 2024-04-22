// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package grpc

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/evmos/evmos/v17/app"
	"github.com/evmos/evmos/v17/encoding"
)

// GetGrants returns the grants for the given grantee and granter combination.
//
// NOTE: To extract the concrete authorizations, use the GetAuthorizations method.
func (gqh *IntegrationHandler) GetGrants(grantee, granter string) ([]*authz.Grant, error) {
	authzClient := gqh.network.GetAuthzClient()
	res, err := authzClient.Grants(context.Background(), &authz.QueryGrantsRequest{
		Grantee: grantee,
		Granter: granter,
	})
	if err != nil {
		return nil, err
	}

	return res.Grants, nil
}

// GetGrantsByGrantee returns the grants for the given grantee.
//
// NOTE: To extract the concrete authorizations, use the GetAuthorizationsByGrantee method.
func (gqh *IntegrationHandler) GetGrantsByGrantee(grantee string) ([]*authz.GrantAuthorization, error) {
	authzClient := gqh.network.GetAuthzClient()
	res, err := authzClient.GranteeGrants(context.Background(), &authz.QueryGranteeGrantsRequest{
		Grantee: grantee,
	})
	if err != nil {
		return nil, err
	}

	return res.Grants, nil
}

// GetGrantsByGranter returns the grants for the given granter.
//
// NOTE: To extract the concrete authorizations, use the GetAuthorizationsByGranter method.
func (gqh *IntegrationHandler) GetGrantsByGranter(granter string) ([]*authz.GrantAuthorization, error) {
	authzClient := gqh.network.GetAuthzClient()
	res, err := authzClient.GranterGrants(context.Background(), &authz.QueryGranterGrantsRequest{
		Granter: granter,
	})
	if err != nil {
		return nil, err
	}

	return res.Grants, nil
}

// GetAuthorizations returns the concrete authorizations for the given grantee and granter combination.
func (gqh *IntegrationHandler) GetAuthorizations(grantee, granter string) ([]authz.Authorization, error) {
	encodingCfg := encoding.MakeConfig(app.ModuleBasics)

	grants, err := gqh.GetGrants(grantee, granter)
	if err != nil {
		return nil, err
	}

	auths := make([]authz.Authorization, 0, len(grants))
	for _, grant := range grants {
		var auth authz.Authorization
		err := encodingCfg.InterfaceRegistry.UnpackAny(grant.Authorization, &auth)
		if err != nil {
			return nil, err
		}

		auths = append(auths, auth)
	}

	return auths, nil
}

// GetAuthorizationsByGrantee returns the concrete authorizations for the given grantee.
func (gqh *IntegrationHandler) GetAuthorizationsByGrantee(grantee string) ([]authz.Authorization, error) {
	grants, err := gqh.GetGrantsByGrantee(grantee)
	if err != nil {
		return nil, err
	}

	return unpackGrantAuthzs(grants)
}

// GetAuthorizationsByGranter returns the concrete authorizations for the given granter.
func (gqh *IntegrationHandler) GetAuthorizationsByGranter(granter string) ([]authz.Authorization, error) {
	grants, err := gqh.GetGrantsByGranter(granter)
	if err != nil {
		return nil, err
	}

	return unpackGrantAuthzs(grants)
}

// unpackGrantAuthzs unpacks the given grant authorization.
func unpackGrantAuthzs(grantAuthzs []*authz.GrantAuthorization) ([]authz.Authorization, error) {
	encodingCfg := encoding.MakeConfig(app.ModuleBasics)

	auths := make([]authz.Authorization, 0, len(grantAuthzs))
	for _, grantAuthz := range grantAuthzs {
		var auth authz.Authorization
		err := encodingCfg.InterfaceRegistry.UnpackAny(grantAuthz.Authorization, &auth)
		if err != nil {
			return nil, err
		}

		auths = append(auths, auth)
	}

	return auths, nil
}
