package accesscontrol_test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	accesscontrol "github.com/evmos/evmos/v18/precompiles/access_control"
	commonprecompile "github.com/evmos/evmos/v18/precompiles/common"
	"github.com/evmos/evmos/v18/precompiles/testutil"
)

func (s *PrecompileTestSuite) TestHasRole() {
	method := s.precompile.Methods[accesscontrol.MethodHasRole]
	testCases := []struct {
		name     string
		malleate func() []interface{}
		hasRole  []byte
	}{
		{
			"fail - sender does not have the role",
			func() []interface{} {
				return []interface{}{[32]byte(accesscontrol.RoleMinter), s.keyring.GetKey(0).Addr}
			},
			commonprecompile.FalseValue,
		},
		{
			"success - sender has the minter role",
			func() []interface{} {
				s.network.App.AccessControlKeeper.SetRole(s.network.GetContext(), s.precompile.Address(), accesscontrol.RoleMinter, s.keyring.GetKey(0).Addr)
				return []interface{}{[32]byte(accesscontrol.RoleMinter), s.keyring.GetKey(0).Addr}
			},
			commonprecompile.TrueValue,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			_, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), s.keyring.GetKey(0).Addr, s.precompile, 0)
			hasRole, err := s.precompile.HasRole(ctx, &method, tc.malleate())

			s.Require().NoError(err)
			s.Require().Equal(hasRole, tc.hasRole)
		})
	}
}

func (s *PrecompileTestSuite) TestGetRoleAdmin() {
	method := s.precompile.Methods[accesscontrol.MethodGetRoleAdmin]
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expPass     bool
		errContains string
		roleAdmin   common.Hash
	}{
		{
			"fail - invalid role argument",
			func() []interface{} {
				return []interface{}{[]byte("test")}
			},
			false,
			fmt.Sprintf(accesscontrol.ErrInvalidRoleArgument),
			[32]byte{},
		},
		{
			"pass - default admin for new role",
			func() []interface{} {
				return []interface{}{[32]byte(crypto.Keccak256Hash([]byte("DEVELOPER_ROLE")))}
			},
			true,
			"",
			accesscontrol.RoleDefaultAdmin,
		},
		{
			"pass - get the admin role of the minter role",
			func() []interface{} {
				return []interface{}{[32]byte(accesscontrol.RoleMinter)}
			},
			true,
			"",
			accesscontrol.RoleDefaultAdmin,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			_, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), s.keyring.GetKey(0).Addr, s.precompile, 0)
			roleAdmin, err := s.precompile.GetRoleAdmin(ctx, &method, tc.malleate())

			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(roleAdmin, tc.roleAdmin.Bytes())
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestGrantRole() {
	addr := s.keyring.GetKey(0).Addr
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expPass     bool
		errContains string
	}{
		{
			"fail - sender does not have admin role, only admin role can grant roles",
			func() []interface{} {
				return []interface{}{
					[32]uint8(accesscontrol.RoleBurner.Bytes()),
					addr,
				}
			},
			false,
			fmt.Sprintf(accesscontrol.ErrSenderNoRole),
		},
		{
			"pass - user already HAS the role in question",
			func() []interface{} {
				s.network.App.AccessControlKeeper.SetRole(s.network.GetContext(), s.precompile.Address(), accesscontrol.RoleDefaultAdmin, addr)
				s.network.App.AccessControlKeeper.SetRole(s.network.GetContext(), s.precompile.Address(), accesscontrol.RoleBurner, addr)
				return []interface{}{
					[32]uint8(accesscontrol.RoleBurner.Bytes()),
					addr,
				}
			},
			true,
			"",
		},
		{
			"pass - sender has admin role and sets the burner role",
			func() []interface{} {
				s.network.App.AccessControlKeeper.SetRole(s.network.GetContext(), s.precompile.Address(), accesscontrol.RoleDefaultAdmin, addr)
				return []interface{}{
					[32]uint8(accesscontrol.RoleBurner.Bytes()),
					addr,
				}
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			stateDB := s.network.GetStateDB()

			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), addr, s.precompile, 0)
			_, err := s.precompile.GrantRole(ctx, contract, stateDB, tc.malleate())

			if tc.expPass {
				s.Require().NoError(err)
				s.Require().True(s.network.App.AccessControlKeeper.HasRole(ctx, s.precompile.Address(), accesscontrol.RoleBurner, addr))
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestRevokeRole() {
	addr := s.keyring.GetKey(0).Addr
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expPass     bool
		errContains string
	}{
		{
			"fail - sender does not have admin role, only admin role can grant roles",
			func() []interface{} {
				return []interface{}{
					[32]uint8(accesscontrol.RoleBurner.Bytes()),
					addr,
				}
			},
			false,
			fmt.Sprintf(accesscontrol.ErrSenderNoRole),
		},
		{
			"pass - user already does NOT have the role in question",
			func() []interface{} {
				s.network.App.AccessControlKeeper.SetRole(s.network.GetContext(), s.precompile.Address(), accesscontrol.RoleDefaultAdmin, addr)
				return []interface{}{
					[32]uint8(accesscontrol.RoleBurner.Bytes()),
					addr,
				}
			},
			true,
			"",
		},
		{
			"pass - sender has admin role and revokes the burner role",
			func() []interface{} {
				s.network.App.AccessControlKeeper.SetRole(s.network.GetContext(), s.precompile.Address(), accesscontrol.RoleDefaultAdmin, addr)
				s.network.App.AccessControlKeeper.SetRole(s.network.GetContext(), s.precompile.Address(), accesscontrol.RoleBurner, addr)
				return []interface{}{
					[32]uint8(accesscontrol.RoleBurner.Bytes()),
					addr,
				}
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			stateDB := s.network.GetStateDB()

			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), addr, s.precompile, 0)
			_, err := s.precompile.RevokeRole(ctx, contract, stateDB, tc.malleate())

			if tc.expPass {
				s.Require().NoError(err)
				s.Require().False(s.network.App.AccessControlKeeper.HasRole(ctx, s.precompile.Address(), accesscontrol.RoleBurner, addr))
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestRenounceRole() {
	addr := s.keyring.GetKey(0).Addr
	testCases := []struct {
		name        string
		malleate    func() []interface{}
		expPass     bool
		errContains string
	}{
		{
			"fail - sender is not the same as the account",
			func() []interface{} {
				return []interface{}{
					[32]uint8(accesscontrol.RoleBurner.Bytes()),
					s.keyring.GetKey(1).Addr,
				}
			},
			false,
			fmt.Sprintf(accesscontrol.ErrRenounceRoleDifferentThanCaller),
		},
		{
			"pass - user already does NOT have the role in question",
			func() []interface{} {
				return []interface{}{
					[32]uint8(accesscontrol.RoleBurner.Bytes()),
					addr,
				}
			},
			true,
			"",
		},
		{
			"pass - sender is the account and renounces the burner role",
			func() []interface{} {
				s.network.App.AccessControlKeeper.SetRole(s.network.GetContext(), s.precompile.Address(), accesscontrol.RoleBurner, addr)
				return []interface{}{
					[32]uint8(accesscontrol.RoleBurner.Bytes()),
					addr,
				}
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			stateDB := s.network.GetStateDB()

			contract, ctx := testutil.NewPrecompileContract(s.T(), s.network.GetContext(), addr, s.precompile, 0)
			_, err := s.precompile.RenounceRole(ctx, contract, stateDB, tc.malleate())

			if tc.expPass {
				s.Require().NoError(err)
				s.Require().False(s.network.App.AccessControlKeeper.HasRole(ctx, s.precompile.Address(), accesscontrol.RoleBurner, addr))
			} else {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			}
		})
	}
}
