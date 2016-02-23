package main

import "fmt"

func main() {
	var a uint8 = 10
	var b uint8 = 22

	a = a + ^b + 1
	fmt.Printf("v=%d\n", a)

}
