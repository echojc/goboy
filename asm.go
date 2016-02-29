package main

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

func Disassemble(addr uint16) (string, string, uint16) {
	var length uint16 = 1
	op := read(addr)
	opcode := opcodes[op]
	raw := fmt.Sprintf("%02X", op)

	// cb-prefixed code, need to relookup in cbcodes
	if op == 0xcb {
		op = read(addr + 1)
		opcode = cbcodes[op]
		raw += fmt.Sprintf(" %02X", op)
		length += 1
	}

	// bwahaha
	name := runtime.FuncForPC(reflect.ValueOf(opcode).Pointer()).Name()
	parts := strings.Split(name, "_")

	for i := 1; i < len(parts); i++ {
		m := false
		s := false

		if parts[i][0] == 'm' {
			parts[i] = parts[i][1:]
			m = true
		} else if parts[i][0] == 's' {
			parts[i] = parts[i][1:]
			s = true
		}

		if parts[i] == "N" {
			rawV := read(addr + 1)
			v := int16(rawV)
			n := ""
			if s && v >= 0x80 {
				v = 0x100 - v
				n = "-"
			}
			parts[i] = fmt.Sprintf("%s%02xh", n, v)

			length += 1
			raw += fmt.Sprintf(" %02X", rawV)
		} else if parts[i] == "NN" {
			rawH := read(addr + 2)
			rawL := read(addr + 1)
			parts[i] = fmt.Sprintf("%04xh", uint16(rawH)<<8+uint16(rawL))

			length += 2
			raw += fmt.Sprintf(" %02X %02X", rawL, rawH)
		}

		if m {
			parts[i] = "(" + parts[i] + ")"
		}
	}

	prettyPrinted := parts[0][strings.LastIndex(parts[0], ".")+1:] + " " + strings.Join(parts[1:], ", ")
	return raw, prettyPrinted, length
}
