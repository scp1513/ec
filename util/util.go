// Package util 常用工具库
package util

import (
	"crypto/md5"
	"encoding/hex"
	"os"
)

// CStrToString 以\n结尾的字节数组转string
func CStrToString(cstr []byte) string {
	length := len(cstr)
	for i := 0; i < len(cstr); i++ {
		if cstr[i] == 0 {
			length = i
			break
		}
	}
	return string(cstr[:length])
}

func MD5(s ...string) string {
	h := md5.New()
	for i := 0; i < len(s); i++ {
		h.Write([]byte(s[i]))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func MD5File(filepath string) (string, error) {
	var data [4096]byte
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	for {
		n, err := file.Read(data[:])
		if n == 0 && err != nil {
			break
		}
		hash.Write(data[:n])
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
