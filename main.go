package main

import "fmt"

func main() {
	write(0xc000, 0xe8)
	write(0xc001, 0x3a)
	write(0xc002, 0xe8)
	write(0xc003, 0xff)

	var addr uint16 = 0xc000
	for addr < 0xc004 {
		raw, pretty, length := Disassemble(addr)
		fmt.Printf("% -12s%s\n", raw, pretty)
		addr += length
	}
}
