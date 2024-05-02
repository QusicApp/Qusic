//go:build windows

package discordrpc

import (
	"net"
	"time"

	"gopkg.in/natefinch/npipe.v2"
)

func newConnection() (net.Conn, error) {
	sock, err := npipe.DialTimeout(`\\.\pipe\discord-ipc-0`, time.Second*2)
	if err != nil {
		return nil, err
	}
	return sock, nil
}
