// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package upgrade

import "fmt"

// CreateValidatorExec creates a staking tx to create a validator on the chain
func (m *Manager) CreateValidatorExec(targetVersion, chainID string, upgradeHeight uint, legacy bool, feesFlag, gasFlag string) (string, error) {
	// TODO evmosd tx staking create-validator --amount=1000000000000000000aevmos --pubkey=$(evmosd tendermint show-validator) --moniker="Dora Factory" --chain-id="evmos_9000-1" --commission-rate="1.0" --commission-max-rate="1.0" --commission-max-change-rate="1.0" --min-self-delegation="1" --from=dev0 --details="Public Good Staking" --home ~/.tmp-evmosd --fees 10aevmos
	cmd := []string{
		"evmosd",
		"staking",
		"gov",

		fmt.Sprintf("--chain-id=%s", chainID),
		"--from=mykey",
		"-b=block",
		"--yes",
		"--keyring-backend=test",
		"--log_format=json",
		feesFlag,
		gasFlag,
	}
	// increment proposal counter to use proposal number for deposit && voting
	m.proposalCounter++
	return m.CreateExec(cmd, m.ContainerID())
}