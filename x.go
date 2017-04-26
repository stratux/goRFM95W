package main

import (
	"./goRFM95W"
	"fmt"
)

func main() {
	rfm95w, err := goRFM95W.New()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	x, err := rfm95w.GetRegister(0x01)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	fmt.Printf("val=%02x\n", x)

	x, err = rfm95w.SetRegister(0x01, 0x09)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	fmt.Printf("val=%02x\n", x)

}
