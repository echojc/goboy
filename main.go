package main

import "fmt"

func main() {
	var a uint8 = 16
	var b uint8 = 1

	var c = uint8(a) - b
	var h = (uint8(a)^b^c)&0x10 > 0

	fmt.Printf("v=%t\n", h)

}
