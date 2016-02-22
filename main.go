package main

import "fmt"

func main() {
	var a uint8 = 0
	fmt.Printf("%d\n", uint8(int16(a)-1&0xff))
}
