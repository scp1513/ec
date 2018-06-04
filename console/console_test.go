package console

import (
	"log"
	"testing"
)

func TestTree(t *testing.T) {
	AddCmdHandler("add 3 ?u ?s", func(...interface{}) {
		log.Printf("3")
	})
	AddCmdHandler("add 4 ?u ?u", func(...interface{}) {
		log.Printf("4")
	})
	for _, v := range cmds {
		log.Printf("cmd: %v", v)
	}
}

func TestParse(t *testing.T) {
	AddCmdHandler("add 5 ?f64 ?s", func(a ...interface{}) {
		log.Printf("%#v", a)
	})
	parseCmd("add 5 2.1 3")
}
