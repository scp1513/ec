package net

import (
	"net"
	"strings"
)

// tcpServerSession tcp服务端会话
type tcpServerSession struct {
	tcpServer *TCPServer
	conn      net.Conn
	connID    uint
}

func newTCPServerSession(tcpServer *TCPServer, conn net.Conn, connID uint) *tcpServerSession {
	return &tcpServerSession{
		tcpServer: tcpServer,
		conn:      conn,
		connID:    connID,
	}
}

func (t *tcpServerSession) send(data ...[]byte) (int, error) {
	return doSend(t.conn, data...)
}

func (t *tcpServerSession) sendBatch(batch *Batch) (int, error) {
	return t.conn.Write(batch.data())
}

func (t *tcpServerSession) close() {
	t.conn.Close()
}

func (t *tcpServerSession) getIP() string {
	addr := t.conn.RemoteAddr().String()
	ss := strings.Split(addr, ":")
	if len(ss) < 1 {
		return ""
	}
	return ss[0]
}

func (t *tcpServerSession) handleConnection() {
	handleRecv(t.conn, func(data []byte) { t.tcpServer.onRecved(t.connID, data) }, func() { t.tcpServer.onSessionClosed(t.connID) })
}
