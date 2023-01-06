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

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"

	"github.com/hashicorp/go-version"
)

var upgradesPath = "../../app/upgrades"

// Custom comparator for sorting semver version strings
type byVersion []string

func (s byVersion) Len() int { return len(s) }

func (s byVersion) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Compare semver versions strings properly
func (s byVersion) Less(i, j int) bool {
	v1, err := version.NewVersion(s[i])
	if err != nil {
		log.Fatal(err)
	}
	v2, err := version.NewVersion(s[j])
	if err != nil {
		log.Fatal(err)
	}
	return v1.LessThan(v2)
}

// RetrieveUpgradeVersion parses app/upgrades folder and returns slice of semver upgrade versions
// ascending order, e.g ["v1.0.0", "v1.0.1", "v1.1.0", ... , "v10.0.0"]
func (m *Manager) RetrieveUpgradesList() ([]string, error) {
	dirs, err := os.ReadDir(upgradesPath)
	if err != nil {
		return nil, err
	}

	versions := []string{}

	// pattern to find quoted string(upgrade version) in a file e.g. "v10.0.0"
	pattern := regexp.MustCompile(`"(.*?)"`)

	for _, d := range dirs {
		// creating path to upgrade dir file with constant upgrade version
		constantsPath := fmt.Sprintf("%s/%s/constants.go", upgradesPath, d.Name())
		f, err := os.ReadFile(constantsPath)
		if err != nil {
			return nil, err
		}
		v := pattern.FindString(string(f))
		// v[1:len(v)-1] subslice used to remove quotes from version string
		versions = append(versions, v[1:len(v)-1])
	}

	sort.Sort(byVersion(versions))

	return versions, nil
}

// ExportState executes the  'docker cp' command to copy container .evmosd dir
// to the specified target dir (local)
//
// See https://docs.docker.com/engine/reference/commandline/cp/
func (m *Manager) ExportState(targetDir string) error {
	/* #nosec G204 */
	cmd := exec.Command(
		"docker",
		"cp",
		fmt.Sprintf("%s:/root/.evmosd", m.ContainerID()),
		targetDir,
	)
	return cmd.Run()
}
