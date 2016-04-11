package net

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

var (
	sessions  = make(map[uint]*tcpClientSession)
	idCounter uint
	mutex     sync.RWMutex
)

// ConnectTo 客户端主动连接
// @param ip string
// @param port string/int/uint16
// @param onStatus func(uint, error)
// @param onRecved func(uint, []byte)
// @param onClosed func(uint)
func ConnectTo(params map[string]interface{}) uint {
	var (
		ip       string
		port     uint16
		onStatus func(uint, error)
		onRecved func(uint, []byte)
		onClosed func(uint)
	)
	if v, ok := params["port"]; ok {
		switch vv := v.(type) {
		case string:
			i, err := strconv.Atoi(vv)
			if err != nil {
				fmt.Println("invalid port:", vv)
				return 0
			}
			port = uint16(i)
		case int:
			port = uint16(vv)
		case uint16:
			port = vv
		default:
			fmt.Println("invalid ip type:", vv)
			return 0
		}
	}
	if v, ok := params["ip"]; ok {
		if vv, ok := v.(string); ok {
			ip = vv
		}
	}
	if v, ok := params["onStatus"]; ok {
		if vv, ok := v.(func(uint, error)); ok {
			onStatus = vv
		}
	}
	if v, ok := params["onRecved"]; ok {
		if vv, ok := v.(func(uint, []byte)); ok {
			onRecved = vv
		}
	}
	if v, ok := params["onClosed"]; ok {
		if vv, ok := v.(func(uint)); ok {
			onClosed = vv
		}
	}

	mutex.Lock()
	idCounter++
	connID := idCounter
	mutex.Unlock()

	go func(connID uint, ip string, port uint16, onStatus func(uint, error), onRecved func(uint, []byte), onClosed func(uint)) {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
		if err != nil {
			onStatus(connID, err)
			return
		}
		session := newTCPClientSession(conn, connID, onRecved, onClosed)
		mutex.Lock()
		sessions[session.connID] = session
		mutex.Unlock()
		onStatus(connID, nil)
		go session.handleConnection()
	}(connID, ip, port, onStatus, onRecved, onClosed)
	return connID
}

// Disconnect 断开连接
func Disconnect(connID uint) {
	mutex.RLock()
	session, ok := sessions[connID]
	mutex.RUnlock()
	if ok {
		session.close()
	}
}

// Send 发送消息
func Send(connID uint, data ...[]byte) (int, error) {
	mutex.RLock()
	session, ok := sessions[connID]
	mutex.RUnlock()
	if !ok {
		return 0, fmt.Errorf("can't find connID: %d", connID)
	}

	return doSend(session.conn, data...)
}

// SendBatch 发送批量消息
func SendBatch(connID uint, batch *Batch) (int, error) {
	mutex.RLock()
	session, ok := sessions[connID]
	mutex.RUnlock()
	if !ok {
		return 0, fmt.Errorf("can't find connID: %d", connID)
	}

	return session.conn.Write(batch.data())
}

func onSessionClosed(connID uint) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(sessions, connID)
}
