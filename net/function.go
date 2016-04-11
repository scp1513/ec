package net

import (
	"fmt"
	"net"
)

func handleRecv(conn net.Conn, onRecved func([]byte), onClosed func()) {
	for {
		var data0 [3]byte
		_, err := conn.Read(data0[:])
		if err != nil {
			conn.Close()
			onClosed()
			break
		}
		length := uint16(data0[0]) | uint16(data0[1])<<8
		if data0[2] == 0 {
			// |-- packet --|
			data := make([]byte, length)
			_, err = conn.Read(data)
			if err != nil {
				conn.Close()
				onClosed()
				break
			}
			onRecved(data)
		} else {
			// |--num(1)--|--singleLen(2)--|--single packet(n)--|--singleLen(2)--|--single packet(n)--|...
			var numBytes [1]byte
			_, err = conn.Read(numBytes[:])
			if err != nil {
				conn.Close()
				onClosed()
				break
			}

			totalLen := 1
			for i := 0; i < int(numBytes[0]); i++ {
				var lenBytes [2]byte
				_, err = conn.Read(lenBytes[:])
				if err != nil {
					conn.Close()
					onClosed()
					break
				}

				singleLen := uint16(lenBytes[0]) | uint16(lenBytes[1])<<8
				totalLen += int(singleLen) + 2
				data := make([]byte, singleLen)
				_, err = conn.Read(data)
				if err != nil {
					conn.Close()
					onClosed()
					break
				}

				onRecved(data)
			}
			if totalLen != int(length) {
				fmt.Println("totalLen != int(length)")
			}
		}
	}
}

func doSend(conn net.Conn, data ...[]byte) (int, error) {
	var length uint64
	for i := 0; i < len(data); i++ {
		length += uint64(len(data[i]))
	}
	if length >= 0xffff {
		return 0, fmt.Errorf("send data len over limit: %d", length)
	}

	d := make([]byte, length+3)
	d[0] = byte(length & 0xff)
	d[1] = byte(length >> 8)
	d[2] = 0
	index := 3
	for i := 0; i < len(data); i++ {
		copy(d[index:], data[i])
		index += len(data[i])
	}
	conn.Write(d)
	return int(length), nil
}
