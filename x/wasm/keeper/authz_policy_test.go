package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/evoblockchain/evoblock/x/wasm/types"
)

func TestDefaultAuthzPolicyCanCreateCode(t *testing.T) {
	myActorAddress := RandomAccountAddress(t)
	otherAddress := RandomAccountAddress(t)
	specs := map[string]struct {
		chainConfigs     ChainAccessConfigs
		contractInstConf types.AccessConfig
		actor            sdk.AccAddress
		exp              bool
		panics           bool
	}{
		"upload nobody": {
			chainConfigs:     NewChainAccessConfigs(types.AllowNobody, types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
			exp:              false,
		},
		"upload everybody": {
			chainConfigs:     NewChainAccessConfigs(types.AllowEverybody, types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
			exp:              true,
		},
		"upload only address - same": {
			chainConfigs:     NewChainAccessConfigs(types.AccessTypeOnlyAddress.With(myActorAddress), types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
			exp:              true,
		},
		"upload only address - different": {
			chainConfigs:     NewChainAccessConfigs(types.AccessTypeOnlyAddress.With(otherAddress), types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
			exp:              false,
		},
		"upload any address - included": {
			chainConfigs:     NewChainAccessConfigs(types.AccessTypeAnyOfAddresses.With(otherAddress, myActorAddress), types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
			exp:              true,
		},
		"upload any address - not included": {
			chainConfigs:     NewChainAccessConfigs(types.AccessTypeAnyOfAddresses.With(otherAddress), types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
			exp:              false,
		},
		"contract config -  subtype": {
			chainConfigs:     NewChainAccessConfigs(types.AllowEverybody, types.AllowEverybody),
			contractInstConf: types.AccessTypeAnyOfAddresses.With(myActorAddress),
			exp:              true,
		},
		"contract config - not subtype": {
			chainConfigs:     NewChainAccessConfigs(types.AllowEverybody, types.AllowNobody),
			contractInstConf: types.AllowEverybody,
			exp:              false,
		},
		"upload undefined config - panics": {
			chainConfigs:     NewChainAccessConfigs(types.AccessConfig{}, types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
			panics:           true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			policy := DefaultAuthorizationPolicy{}
			if !spec.panics {
				got := policy.CanCreateCode(spec.chainConfigs, myActorAddress, spec.contractInstConf)
				assert.Equal(t, spec.exp, got)
				return
			}
			assert.Panics(t, func() {
				policy.CanCreateCode(spec.chainConfigs, myActorAddress, spec.contractInstConf)
			})
		})
	}
}

func TestDefaultAuthzPolicyCanInstantiateContract(t *testing.T) {
	myActorAddress := RandomAccountAddress(t)
	otherAddress := RandomAccountAddress(t)
	specs := map[string]struct {
		config types.AccessConfig
		actor  sdk.AccAddress
		exp    bool
		panics bool
	}{
		"nobody": {
			config: types.AllowNobody,
			exp:    false,
		},
		"everybody": {
			config: types.AllowEverybody,
			exp:    true,
		},
		"only address - same": {
			config: types.AccessTypeOnlyAddress.With(myActorAddress),
			exp:    true,
		},
		"only address - different": {
			config: types.AccessTypeOnlyAddress.With(otherAddress),
			exp:    false,
		},
		"any address - included": {
			config: types.AccessTypeAnyOfAddresses.With(otherAddress, myActorAddress),
			exp:    true,
		},
		"any address - not included": {
			config: types.AccessTypeAnyOfAddresses.With(otherAddress),
			exp:    false,
		},
		"undefined config - panics": {
			config: types.AccessConfig{},
			panics: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			policy := DefaultAuthorizationPolicy{}
			if !spec.panics {
				got := policy.CanInstantiateContract(spec.config, myActorAddress)
				assert.Equal(t, spec.exp, got)
				return
			}
			assert.Panics(t, func() {
				policy.CanInstantiateContract(spec.config, myActorAddress)
			})
		})
	}
}

func TestDefaultAuthzPolicyCanModifyContract(t *testing.T) {
	myActorAddress := RandomAccountAddress(t)
	otherAddress := RandomAccountAddress(t)

	specs := map[string]struct {
		admin sdk.AccAddress
		exp   bool
	}{
		"same as actor": {
			admin: myActorAddress,
			exp:   true,
		},
		"different admin": {
			admin: otherAddress,
			exp:   false,
		},
		"no admin": {
			exp: false,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			policy := DefaultAuthorizationPolicy{}
			got := policy.CanModifyContract(spec.admin, myActorAddress)
			assert.Equal(t, spec.exp, got)
		})
	}
}

func TestDefaultAuthzPolicyCanModifyCodeAccessConfig(t *testing.T) {
	myActorAddress := RandomAccountAddress(t)
	otherAddress := RandomAccountAddress(t)

	specs := map[string]struct {
		admin  sdk.AccAddress
		subset bool
		exp    bool
	}{
		"same as actor - subset": {
			admin:  myActorAddress,
			subset: true,
			exp:    true,
		},
		"same as actor - not subset": {
			admin:  myActorAddress,
			subset: false,
			exp:    false,
		},
		"different admin": {
			admin: otherAddress,
			exp:   false,
		},
		"no admin": {
			exp: false,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			policy := DefaultAuthorizationPolicy{}
			got := policy.CanModifyCodeAccessConfig(spec.admin, myActorAddress, spec.subset)
			assert.Equal(t, spec.exp, got)
		})
	}
}

func TestGovAuthzPolicyCanCreateCode(t *testing.T) {
	myActorAddress := RandomAccountAddress(t)
	otherAddress := RandomAccountAddress(t)
	specs := map[string]struct {
		chainConfigs     ChainAccessConfigs
		contractInstConf types.AccessConfig
		actor            sdk.AccAddress
	}{
		"upload nobody": {
			chainConfigs:     NewChainAccessConfigs(types.AllowNobody, types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
		},
		"upload everybody": {
			chainConfigs:     NewChainAccessConfigs(types.AllowEverybody, types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
		},
		"upload only address - same": {
			chainConfigs:     NewChainAccessConfigs(types.AccessTypeOnlyAddress.With(myActorAddress), types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
		},
		"upload only address - different": {
			chainConfigs:     NewChainAccessConfigs(types.AccessTypeOnlyAddress.With(otherAddress), types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
		},
		"upload any address - included": {
			chainConfigs:     NewChainAccessConfigs(types.AccessTypeAnyOfAddresses.With(otherAddress, myActorAddress), types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
		},
		"upload any address - not included": {
			chainConfigs:     NewChainAccessConfigs(types.AccessTypeAnyOfAddresses.With(otherAddress), types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
		},
		"contract config -  subtype": {
			chainConfigs:     NewChainAccessConfigs(types.AllowEverybody, types.AllowEverybody),
			contractInstConf: types.AccessTypeAnyOfAddresses.With(myActorAddress),
		},
		"contract config - not subtype": {
			chainConfigs:     NewChainAccessConfigs(types.AllowEverybody, types.AllowNobody),
			contractInstConf: types.AllowEverybody,
		},
		"upload undefined config - not panics": {
			chainConfigs:     NewChainAccessConfigs(types.AccessConfig{}, types.AllowEverybody),
			contractInstConf: types.AllowEverybody,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			policy := GovAuthorizationPolicy{}
			got := policy.CanCreateCode(spec.chainConfigs, myActorAddress, spec.contractInstConf)
			assert.True(t, got)
		})
	}
}

func TestGovAuthzPolicyCanInstantiateContract(t *testing.T) {
	myActorAddress := RandomAccountAddress(t)
	otherAddress := RandomAccountAddress(t)
	specs := map[string]struct {
		config types.AccessConfig
		actor  sdk.AccAddress
	}{
		"nobody": {
			config: types.AllowNobody,
		},
		"everybody": {
			config: types.AllowEverybody,
		},
		"only address - same": {
			config: types.AccessTypeOnlyAddress.With(myActorAddress),
		},
		"only address - different": {
			config: types.AccessTypeOnlyAddress.With(otherAddress),
		},
		"any address - included": {
			config: types.AccessTypeAnyOfAddresses.With(otherAddress, myActorAddress),
		},
		"any address - not included": {
			config: types.AccessTypeAnyOfAddresses.With(otherAddress),
		},
		"undefined config - panics": {
			config: types.AccessConfig{},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			policy := GovAuthorizationPolicy{}
			got := policy.CanInstantiateContract(spec.config, myActorAddress)
			assert.True(t, got)
		})
	}
}

func TestGovAuthzPolicyCanModifyContract(t *testing.T) {
	myActorAddress := RandomAccountAddress(t)
	otherAddress := RandomAccountAddress(t)

	specs := map[string]struct {
		admin sdk.AccAddress
	}{
		"same as actor": {
			admin: myActorAddress,
		},
		"different admin": {
			admin: otherAddress,
		},
		"no admin": {},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			policy := GovAuthorizationPolicy{}
			got := policy.CanModifyContract(spec.admin, myActorAddress)
			assert.True(t, got)
		})
	}
}

func TestGovAuthzPolicyCanModifyCodeAccessConfig(t *testing.T) {
	myActorAddress := RandomAccountAddress(t)
	otherAddress := RandomAccountAddress(t)

	specs := map[string]struct {
		admin  sdk.AccAddress
		subset bool
	}{
		"same as actor - subset": {
			admin:  myActorAddress,
			subset: true,
		},
		"same as actor - not subset": {
			admin:  myActorAddress,
			subset: false,
		},
		"different admin": {
			admin: otherAddress,
		},
		"no admin": {},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			policy := GovAuthorizationPolicy{}
			got := policy.CanModifyCodeAccessConfig(spec.admin, myActorAddress, spec.subset)
			assert.True(t, got)
		})
	}
}
