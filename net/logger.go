package net

import (
	"code.google.com/p/log4go"
)

var (
	log log4go.Logger
)

func SetLogger(l log4go.Logger) {
	log = l
}
