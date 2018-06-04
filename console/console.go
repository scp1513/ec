package console

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Handler func(...interface{})

var (
	cmds = make(map[string]H)
)

type H struct {
	s       string
	params  []string
	handler Handler
}

func Start() {
	go startPolling()
}

// ? or ?s : string
// ?b: bool
// ?u8 ~ ?u64: uint8 ~ uint64
// ?s8 ~ ?s64: int8 ~ int64
// ?f32, f64: float32, float64
// eg: "add baller ?u16 ?s"
func AddCmdHandler(cmd string, handler Handler) {
	if handler == nil {
		panic("handler can not nil")
	}

	h := H{handler: handler}

	scaner := bufio.NewScanner(strings.NewReader(strings.ToLower(cmd)))
	scaner.Split(bufio.ScanWords)
	for scaner.Scan() {
		word := scaner.Text()
		if word[0] == '?' {
			h.params = append(h.params, word)
		} else {
			h.s += " "
			h.s += word
		}
	}

	if h.s == "" {
		panic(fmt.Sprintf("cmd format invalid %s", cmd))
	}

	_, ok := cmds[h.s]
	if ok {
		panic(fmt.Sprintf("cmd already exists %s", cmd))
	}

	cmds[h.s] = h
}

func startPolling() {
	reader := bufio.NewReader(os.Stdin)
	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		parseCmd(cmd)
	}
}

func parseCmd(cmd string) {
	var s string
	var v *H

	scaner := bufio.NewScanner(strings.NewReader(strings.ToLower(cmd)))
	scaner.Split(bufio.ScanWords)
	for scaner.Scan() {
		word := scaner.Text()
		s += " "
		s += word
		if h, ok := cmds[s]; ok {
			v = &h
			break
		}
	}

	if v == nil {
		return
	}

	params := make([]interface{}, len(v.params))
	index := 0
	for scaner.Scan() {
		word := scaner.Text()
		p := parseParam(v.params[index], word)
		if p == nil {
			return
		}
		params[index] = p
		index++
	}

	if len(params) != len(v.params) {
		return
	}

	v.handler(params...)
}

func parseParam(s, v string) interface{} {
	switch s {
	case "?", "?s":
		return v
	case "?b":
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil
		}
		return b
	case "?u8":
		v, err := strconv.ParseUint(v, 0, 8)
		if err != nil {
			return nil
		}
		return uint8(v)
	case "?u16":
		v, err := strconv.ParseUint(v, 0, 16)
		if err != nil {
			return nil
		}
		return uint16(v)
	case "?u32":
		v, err := strconv.ParseUint(v, 0, 32)
		if err != nil {
			return nil
		}
		return uint32(v)
	case "?u64":
		v, err := strconv.ParseUint(v, 0, 64)
		if err != nil {
			return nil
		}
		return uint64(v)
	case "?s8":
		v, err := strconv.ParseInt(v, 0, 8)
		if err != nil {
			return nil
		}
		return int8(v)
	case "?s16":
		v, err := strconv.ParseInt(v, 0, 16)
		if err != nil {
			return nil
		}
		return int16(v)
	case "?s32":
		v, err := strconv.ParseInt(v, 0, 32)
		if err != nil {
			return nil
		}
		return int32(v)
	case "?s64":
		v, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			return nil
		}
		return int64(v)
	case "?f32":
		v, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return nil
		}
		return float32(v)
	case "?f64":
		v, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil
		}
		return v
	}
	return nil
}
