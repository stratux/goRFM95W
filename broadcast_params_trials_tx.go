package main

import (
	"./goRFM95W"
	"fmt"
	"time"
)

const (
	TEST_FREQ = 915000000
)

var testParams = []goRFM95W.RFM95W_Params{
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 12, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 11, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},

	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 11, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},

	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},

	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},

	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},
}

func main() {
	rfm95w, err := goRFM95W.New()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}

	rfm95w.Start()

	rfm95w.Debug = true

	buf := make([]byte, 96)
	for i := 0; i < 96; i++ {
		buf[i] = byte(i + 33)
	}

	for {
		for _, param := range testParams {
			fmt.Printf("%v\n", param)
			err := rfm95w.SetParams(param)
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
			}

			err = rfm95w.SendSync(buf)
			if err != nil {
				fmt.Printf("SendSync() err: %s\n", err.Error())
			}
			fmt.Printf("done\n")
			time.Sleep(100 * time.Millisecond)
		}
	}
}
