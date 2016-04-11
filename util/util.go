// Package util 常用工具库
package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
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
