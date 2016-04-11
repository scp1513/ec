package net

import (
	"net"
)

// tcpClientSession tcp客户端会话
type tcpClientSession struct {
	conn     net.Conn
	connID   uint
	onRecved func(uint, []byte)
	onClosed func(uint)
}

func newTCPClientSession(conn net.Conn, connID uint, onRecved func(uint, []byte), onClosed func(uint)) *tcpClientSession {
	return &tcpClientSession{
		conn:     conn,
		connID:   connID,
		onRecved: onRecved,
		onClosed: onClosed,
	}
}

func (t *tcpClientSession) close() {
	t.conn.Close()
}

func (t *tcpClientSession) handleConnection() {
	handleRecv(t.conn, func(data []byte) { t.onRecved(t.connID, data) }, func() { onSessionClosed(t.connID); t.onClosed(t.connID) })
}
