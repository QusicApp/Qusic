package discordrpc

import (
	"net"
	"os"
	"runtime"
	"time"

	"gopkg.in/natefinch/npipe.v2"
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
	if runtime.GOOS == "windows" {
		sock, err := npipe.DialTimeout(`\\.\pipe\discord-ipc-0`, time.Second*2)
		if err != nil {
			return nil, err
		}
		return sock, nil
	} else {
		sock, err := net.DialTimeout("unix", getIpcPath()+"/discord-ipc-0", time.Second*2)
		if err != nil {
			return nil, err
		}

		return sock, nil
	}
}
