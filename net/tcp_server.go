package net

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

// TCPServer tcp服务端
type TCPServer struct {
	listening   bool
	listener    net.Listener
	sessions    map[uint]*tcpServerSession
	idCounter   uint
	ip          string
	port        uint16
	onConnected func(uint)
	onRecved    func(uint, []byte)
	onClosed    func(uint)
	mutex       sync.RWMutex
}

// NewTCPServer 创建tcpserver
// param [ip] string
// param [port] string/int/uint16
// param [onConnected] func(uint)
// param [onRecved] func(uint, []byte)
// param [onClosed] func(uint)
func NewTCPServer(params map[string]interface{}) (*TCPServer, error) {
	t := &TCPServer{
		listening: false,
		sessions:  make(map[uint]*tcpServerSession),
		idCounter: 0,
	}
	if v, ok := params["port"]; ok {
		switch vv := v.(type) {
		case string:
			i, err := strconv.Atoi(vv)
			if err != nil {
				return nil, fmt.Errorf("invalid ip: %s", vv)
			}
			t.port = uint16(i)
		case int:
			t.port = uint16(vv)
		case uint16:
			t.port = vv
		default:
			return nil, fmt.Errorf("invalid ip type: %#v", vv)
		}
	}
	if v, ok := params["ip"]; ok {
		if vv, ok := v.(string); ok {
			t.ip = vv
		}
	}
	if v, ok := params["onConnected"]; ok {
		if vv, ok := v.(func(uint)); ok {
			t.onConnected = vv
		}
	}
	if v, ok := params["onRecved"]; ok {
		if vv, ok := v.(func(uint, []byte)); ok {
			t.onRecved = vv
		}
	}
	if v, ok := params["onClosed"]; ok {
		if vv, ok := v.(func(uint)); ok {
			t.onClosed = vv
		}
	}
	if t.ip == "" || t.port == 0 || t.onConnected == nil || t.onRecved == nil || t.onClosed == nil {
		return nil, fmt.Errorf("invalid param")
	}
	return t, nil
}

// Start 开始接受连接
func (t *TCPServer) Start() {
	go t.Run()
}

// Run 开始接受连接
func (t *TCPServer) Run() error {
	if t.listening {
		return fmt.Errorf("already start")
	}
	t.listening = true
	return t.startAccept()
}

// Stop 关闭所有连接并停止监听
func (t *TCPServer) Stop() {
	if !t.listening {
		return
	}
	t.listening = false
	t.listener.Close()
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	for _, session := range t.sessions {
		session.close()
	}
}

// StopListen 停止监听
func (t *TCPServer) StopListen() {
	if !t.listening {
		return
	}
	t.listening = false
	t.listener.Close()
}

// CloseAll 关闭所有连接
func (t *TCPServer) CloseAll() {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	for _, session := range t.sessions {
		session.close()
	}
}

// Send 发送消息
func (t *TCPServer) Send(connID uint, data ...[]byte) (int, error) {
	t.mutex.RLock()
	session, ok := t.sessions[connID]
	t.mutex.RUnlock()
	if !ok {
		return 0, fmt.Errorf("can't find connID: %d", connID)
	}

	return session.send(data...)
}

// SendBatch 批量发送
func (t *TCPServer) SendBatch(connID uint, batch *Batch) (int, error) {
	t.mutex.RLock()
	session, ok := t.sessions[connID]
	t.mutex.RUnlock()
	if !ok {
		return 0, fmt.Errorf("can't find connID: %d", connID)
	}

	return session.sendBatch(batch)
}

// Close 关闭指定连接
func (t *TCPServer) Close(connID uint) {
	t.mutex.RLock()
	session, ok := t.sessions[connID]
	t.mutex.RUnlock()
	if !ok {
		return
	}
	session.close()
}

// GetIP 获取指定连接的IP
func (t *TCPServer) GetIP(connID uint) (string, error) {
	t.mutex.RLock()
	session, ok := t.sessions[connID]
	t.mutex.RUnlock()
	if !ok {
		return "", fmt.Errorf("can't find connID: %d", connID)
	}
	return session.getIP(), nil
}

// GetConnCount 获取连接数量
func (t *TCPServer) GetConnCount() uint {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return uint(len(t.sessions))
}

func (t *TCPServer) onSessionClosed(connID uint) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, ok := t.sessions[connID]; ok {
		t.onClosed(connID)
		delete(t.sessions, connID)
	}
}

func (t *TCPServer) startAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", t.ip, t.port))
	if err != nil {
		return err
	}

	for t.listening {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		t.idCounter++
		session := newTCPServerSession(t, conn, t.idCounter)
		t.mutex.Lock()
		if !t.listening {
			t.mutex.Unlock()
			break
		}
		t.sessions[t.idCounter] = session
		t.mutex.Unlock()

		t.onConnected(session.connID)

		go session.handleConnection()
	}
	return nil
}
