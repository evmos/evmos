package utils

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var muPort = &sync.RWMutex{}
var curPort = 10_000

const maxPort = 65535

// GetNextPortAvailable check if port is available and return it. If not available, it will check the next port.
func GetNextPortAvailable() int {
	muPort.Lock()
	defer muPort.Unlock()

	retry := 1000

	for {
		curPort++
		retry--
		if retry < 1 {
			panic("too many retries allocating port")
		}
		if curPort > maxPort {
			curPort = 10_000
		}
		closed, err := isPortClosed(curPort)
		if err != nil {
			continue
		}
		if closed {
			return curPort
		}
	}
}

// isPortClosed checks if given port is closed.
func isPortClosed(port int) (closed bool, err error) {
	closed = true // default: treating as closed

	var conn net.Conn

	conn, err = net.DialTimeout("tcp", net.JoinHostPort("localhost", fmt.Sprintf("%d", port)), time.Second)
	if err != nil {
		errMsg := err.Error()
		dialingError := strings.Contains(errMsg, "error while dialing") || strings.Contains(errMsg, "connection refused")
		if dialingError {
			closed = true
			err = nil
		} else {
			closed = false
		}
	} else {
		if conn != nil {
			defer func() {
				_ = conn.Close()
			}()
			closed = false
		} else {
			closed = false
		}
	}

	return
}
