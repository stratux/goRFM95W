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

	rfm95w.Start()

	buf := make([]byte, 96)
	for i := 0; i < 96; i++ {
		buf[i] = byte(i + 33)
	}
	buf = append([]byte("KE8GRB|"), buf...)

	for {
		for i, param := range testParams {
			fmt.Printf("%d %v\n", i, param)
			err := rfm95w.SetParams(param)
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
			}

			err = rfm95w.SendSync(buf)
			if err != nil {
				fmt.Printf("SendSync() err: %s\n", err.Error())
			}
			fmt.Printf("done, %d ms\n", rfm95w.LastTXTime/time.Millisecond)
			time.Sleep(150 * time.Millisecond)
		}
	}
}
