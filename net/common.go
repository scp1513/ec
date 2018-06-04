package net

import (
	"net"
	"sync"
)

var (
	idCounterMutex sync.Mutex
	idCounter      uint
)

func handleRecv(ip string, connID uint, handler OnRecvedHandler, buf []byte) int {
	for {
		if len(buf) <= 4 {
			return len(buf)
		}

		var (
			// length = packLen - sizeof(length) - sizeof(label)
			length  = uint16(buf[0]) | uint16(buf[1])<<8
			label   = uint16(buf[2]) | uint16(buf[3])<<8
			packLen = 4 + int(length)
		)

		if len(buf) < packLen {
			return len(buf)
		}

		if label == 24 {
			// |--label(2)--|-- packet --|
			packet := make([]byte, length)
			copy(packet, buf[4:4+length])
			handler(connID, packet)
			log.Debug("%s recv packet, length: %d", ip, length)
		} else {
			// |--label(2)--|--num(1)--|--singleLen(2)--|--single packet...--|
			if len(buf) <= 5 {
				return len(buf)
			}

			var (
				num   = int(buf[4])
				index = 5
			)

			for i := 0; i < num; i++ {
				if len(buf) <= index+2 {
					return len(buf)
				}

				singleLen := uint16(buf[index]) | uint16(buf[index+1])<<8
				index += 2

				if len(buf) < index+int(singleLen) {
					return len(buf)
				}

				packet := make([]byte, singleLen)
				copy(packet, buf[index:index+int(singleLen)])
				handler(connID, packet)
				log.Debug("%s recv packet, length: %d", ip, singleLen)
				index += int(singleLen)
			}
		}

		// 将后面的数据移到前面
		for i := 0; i < len(buf)-packLen; i++ {
			buf[i] = buf[i+packLen]
		}

		buf = buf[:len(buf)-packLen]
	}
}

func doSend(ip string, connID uint, conn net.Conn, data []byte) int {
	var (
		length uint16 = uint16(len(data))
		label  uint16 = 24
	)

	var packetInfoBytes [4]byte
	packetInfoBytes[0] = byte(length)
	packetInfoBytes[1] = byte(length >> 8)
	packetInfoBytes[2] = byte(label)
	packetInfoBytes[3] = byte(label >> 8)

	// 发送包长
	ret, err := conn.Write(packetInfoBytes[:])
	if err != nil {
		log.Error("%s: %s", ip, err.Error())
		return 0
	}

	// 发送包体
	ret, err = conn.Write(data)
	if err != nil {
		log.Error("%s: %s", ip, err.Error())
		return 0
	}
	log.Debug("%s send packet, length: %d", ip, length)

	return ret
}
