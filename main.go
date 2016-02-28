package main

import "fmt"

func main() {
	var sp uint16 = 0x12f4
	var n uint8 = 0x10

	c := int32(sp)
	d := int32(n)

	fmt.Printf("v=0x%08x\n", c^d^(c+d))
}
