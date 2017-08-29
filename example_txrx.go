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

	for {
		rfm95w.Send(buf)
		time.Sleep(1 * time.Second)
		msgs := rfm95w.FlushRXBuffer()
		if len(msgs) > 0 {
			fmt.Printf("%d messages received:\n", len(msgs))
			for _, msg := range msgs {
				fmt.Printf("********** Msg Received: RSSI=%d, SNR=%f **********\n", msg.RSSI, msg.SNR)
				fmt.Printf("Frequency: %0.3f MHz, Bandwidth: %d kHz, SpreadingFactor: %d, CodingRate: %d, PremableLength: %d\n", float32(msg.Params.Frequency)/float32(1000000.0), msg.Params.Bandwidth/1000, msg.Params.SpreadingFactor, msg.Params.CodingRate, msg.Params.PreambleLength)
				fmt.Printf("%s: %s\n", msg.Received, string(msg.Buf))
				fmt.Printf("**********\n")
			}
		}
	}
}
