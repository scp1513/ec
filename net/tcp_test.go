package net

import (
	"testing"
)

func TestTcp(t *testing.T) {
	tcpsrv := NewTcpServer(&TcpServerParams{
		Ip:          "127.0.0.1",
		Port:        35180,
		OnConnected: func(uint) {},
		OnRecved:    func(uint, []byte) {},
		OnClosed:    func(uint) {},
	})
	tcpsrv.Start()

	tcpsrv.
}
