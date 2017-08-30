package main

import (
	"./goRFM95W"
	"flag"
	"fmt"
	"time"
)

func main() {
	rfm95w, err := goRFM95W.New(nil)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}

	rfm95w.Start()

	// Generate "test message".
	buf := make([]byte, 96)
	for i := 0; i < 96; i++ {
		buf[i] = byte(i + 33)
	}

	txMode := flag.Bool("tx", false, "TX a message every ~1sec.")
	flag.Parse()

	if *txMode {
		fmt.Printf("TX+RX\n")
	} else {
		fmt.Printf("RX only\n")
	}

	for {
		if *txMode {
			rfm95w.Send(buf)
		}
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
