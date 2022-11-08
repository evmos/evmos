package upgrade

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func (m *Manager) RetrieveUpgradeVersion() (string, error) {
	dirs, err := os.ReadDir("./app/upgrades")
	if err != nil {
		return "", err
	}

	var highest int
	var version string

	pattern := regexp.MustCompile(`"(.*?)"`)
	for _, d := range dirs {
		f, err := os.ReadFile("./app/upgrades/" + d.Name() + "/" + "constants.go")
		if err != nil {
			return "", err
		}
		v := pattern.FindString(string(f))
		numeric := strings.ReplaceAll(v[2:len(v)-1], ".", "")
		number, err := strconv.Atoi(numeric)
		if err != nil {
			return "", err
		}
		if highest < number {
			highest = number
			version = v
		}
	}
	return version, nil
}

// ExportState execute 'docker cp' command to copy container .evmosd dir
// to specified local dir
// https://docs.docker.com/engine/reference/commandline/cp/
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
