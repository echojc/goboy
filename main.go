package main

import "fmt"

func main() {
	var a uint8 = 234
	var b uint8 = 234

	sum := int16(a) + ^int16(b)
	var trunk uint8 = uint8(sum)

	var fh = (a^b^trunk)&0x10 > 0

	fmt.Printf("v=%d z=%t n=%t h=%t c=%t\n", trunk, trunk == 0, true, fh, sum < 0)
}
