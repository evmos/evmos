// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// The config package provides a convinient way to modify x/evm params and values.
// Its primary purpose is to be used during application initialization.

package config

import (
	"github.com/evmos/evmos/v20/x/evm/types"
)

var chainConfig types.ChainConfig = types.DefaultChainConfig()

// SetChainConfig allows to set the chain configuration variable modifying the
// default values.
func SetChainConfig(cc types.ChainConfig) error {
	if err := cc.Validate(); err != nil {
		return err
	}
	chainConfig = cc
	return nil
}

// GetChainConfig returns the chain configuration used in the EVM.
func GetChainConfig() types.ChainConfig {
	return chainConfig
}
