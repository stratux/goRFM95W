package main

import (
	"./goRFM95W"
	"fmt"
	"time"
)

const (
	TEST_FREQ = 915000000
)

type MySituation struct {
	Lat     float32
	Lng     float32
	Alt     float32 // Feet MSL
	GPSTime time.Time
}

var Location MySituation

var testParams = []goRFM95W.RFM95W_Params{
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 12, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 11, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 6, CodingRate: 5, PreambleLength: 8},

	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 11, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 6, CodingRate: 5, PreambleLength: 8},

	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 6, CodingRate: 5, PreambleLength: 8},

	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 6, CodingRate: 5, PreambleLength: 8},

	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 6, CodingRate: 5, PreambleLength: 8},
}

func main() {
	rfm95w, err := goRFM95W.New()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}

	rfm95w.Start()

	rfm95w.Debug = true

	for {
		for _, param := range testParams {
			fmt.Printf("%v\n", param)
			err := rfm95w.SetParams(param)
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
			}

			for {
				msgs := rfm95w.FlushRXBuffer()
				if len(msgs) > 0 {
					for _, msg := range msgs {
						fmt.Printf("%f,%f,%f,%s,%d dBm,%f dB,%0.3f MHz,%0.3f kHz,%d,%d,%d,%s\n", Location.Lat, Location.Lng, Location.Alt, Location.GPSTime, msg.RSSI, msg.SNR, float32(msg.Params.Frequency)/float32(1000000.0), float32(msg.Params.Bandwidth)/float32(1000.0), msg.Params.SpreadingFactor, msg.Params.CodingRate, msg.Params.PreambleLength, msg.Buf)
					}
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}
