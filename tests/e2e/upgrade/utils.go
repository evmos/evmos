package upgrade

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var upgradesPath = "../../app/upgrades"

// RetrieveUpgradeVersion parses the latest upgrade version from the app/upgrades folder
func (m *Manager) RetrieveUpgradeVersion() (string, error) {
	dirs, err := os.ReadDir(upgradesPath)
	if err != nil {
		return "", err
	}
	var highest int
	var version string

	// pattern to find quoted string(upgrade version) in a file e.g. "v10.0.0"
	pattern := regexp.MustCompile(`"(.*?)"`)
	for _, d := range dirs {
		constantsPath := fmt.Sprintf("%s/%s/constants.go", upgradesPath, d.Name())
		f, err := os.ReadFile(constantsPath)
		if err != nil {
			return "", err
		}

		v := pattern.FindString(string(f))

		// [2:len(v)-1] index to remove quotes and 'v' prefix e.g. "v10.0.0" --> 10.0.0
		// than removing dots '.'
		numeric := strings.ReplaceAll(v[2:len(v)-1], ".", "")
		// string: 10000 --> int: 10000
		number, err := strconv.Atoi(numeric)
		if err != nil {
			return "", err
		}

		if highest < number {
			highest = number
			version = v[1 : len(v)-1]
		}
	}
	return version, nil
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
