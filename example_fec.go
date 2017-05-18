package main

import (
	"./goRFM95W"
	"fmt"
	"time"
)

func main() {
	rfm95w, err := goRFM95W.New()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}

	x, err := rfm95w.GetRegister(0x1D)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	fmt.Printf("reg 0x1D=%02x\n", x)

	x, err = rfm95w.GetRegister(0x1E)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	fmt.Printf("reg 0x1E=%02x\n", x)

	x, err = rfm95w.GetRegister(0x26)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	fmt.Printf("reg 0x26=%02x\n", x)

	rfm95w.Start()

	buf := make([]byte, 200)
	for i := 0; i < 200; i++ {
		buf[i] = byte(i)
	}

	buf_encoded := goRFM95W.PacketEncode(buf)

	for {
		time.Sleep(1 * time.Second)
		rfm95w.Send(buf_encoded)
		msgs := rfm95w.FlushRXBuffer()
		if len(msgs) > 0 {
			fmt.Printf("%d messages received:\n", len(msgs))
			for _, msg := range msgs {
				if len(msg.Buf) != 255 {
					fmt.Printf("SKIPPING! Message too short. len(msg.Buf)=%d.\n", len(msg.Buf))
					continue
				}
				msg_corrected, _ := goRFM95W.PacketEncode(msg.Buf)
				fmt.Printf("Message: ")
				for i := 0; i < len(msg_corrected); i++ {
					fmt.Printf("%02x ", msg_corrected[i])
				}
				fmt.Printf("\n")
			}
		}
	}
}
