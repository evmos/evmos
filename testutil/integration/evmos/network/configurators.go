// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package network

import (
	"github.com/evmos/evmos/v20/app"
)

func Test18DecimalsAppConfigurator(chainID string) error {
	return app.Configurator(chainID)
}

func Test6DecimalsAppConfigurator(chainID string) error {
	return app.Configurator(chainID)
}
