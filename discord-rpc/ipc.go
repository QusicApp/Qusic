//go:build !windows

package discordrpc

import (
	"net"
	"os"
	"time"
)

func getIpcPath() string {
	variablesnames := []string{"XDG_RUNTIME_DIR", "TMPDIR", "TMP", "TEMP"}

	for _, variablename := range variablesnames {
		path, exists := os.LookupEnv(variablename)

		if exists {
			return path
		}
	}

	return "/tmp"
}

func newConnection() (net.Conn, error) {
	sock, err := net.DialTimeout("unix", getIpcPath()+"/discord-ipc-0", time.Second*2)
	if err != nil {
		return nil, err
	}

	return sock, nil
}
