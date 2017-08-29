package main

import (
	"./goRFM95W"
	"time"
)

const (
	TEST_FREQ = 915000000
)

var testParams = []goRFM95W.RFM95W_Params{
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},  // -141.5 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},   // -138.8 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},   // -136.1 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8},  // -135 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 15600, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},   // -132.2 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},   // -132 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8}, // -132 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 11, CodingRate: 5, PreambleLength: 8}, // -131.5 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 12, CodingRate: 5, PreambleLength: 8}, // -131 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},   // -129 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},  // -129 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8}, // -129 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 11, CodingRate: 5, PreambleLength: 8}, // -128.5 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 62500, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},   // -126 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 10, CodingRate: 5, PreambleLength: 8}, // -126 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},  // -126 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},  // -126 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 9, CodingRate: 5, PreambleLength: 8},  // -123 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},  // -123 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 125000, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},  // -123 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 8, CodingRate: 5, PreambleLength: 8},  // -120 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 250000, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},  // -120 dBm.
	{Frequency: TEST_FREQ, Bandwidth: 500000, SpreadingFactor: 7, CodingRate: 5, PreambleLength: 8},  // -117 dBm.
}

var testParamsTXTime = []time.Duration{
	8209 * time.Millisecond,
	4432 * time.Millisecond,
	2462 * time.Millisecond,
	2053 * time.Millisecond,
	1395 * time.Millisecond,
	1108 * time.Millisecond,
	1027 * time.Millisecond,
	945 * time.Millisecond,
	863 * time.Millisecond,
	616 * time.Millisecond,
	554 * time.Millisecond,
	513 * time.Millisecond,
	472 * time.Millisecond,
	349 * time.Millisecond,
	257 * time.Millisecond,
	277 * time.Millisecond,
	308 * time.Millisecond,
	139 * time.Millisecond,
	154 * time.Millisecond,
	175 * time.Millisecond,
	77 * time.Millisecond,
	88 * time.Millisecond,
	44 * time.Millisecond,
}
