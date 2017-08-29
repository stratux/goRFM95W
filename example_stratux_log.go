package main

import (
	"./goRFM95W"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	LOGFILE       = "/root/hits.sql"
	SITUATION_URL = "http://192.168.10.1/getSituation"
)

type MySituation struct {
	Lat     float32
	Lng     float32
	Alt     float32 // Feet MSL
	GPSTime time.Time
}

var Location MySituation

var situationMutex *sync.Mutex

func chkErr(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}
}

func situationUpdater() {
	situationUpdateTicker := time.NewTicker(1 * time.Second)
	for {
		<-situationUpdateTicker.C
		situationMutex.Lock()

		resp, err := http.Get(SITUATION_URL)
		if err != nil {
			fmt.Printf("HTTP GET error: %s\n", err.Error())
			situationMutex.Unlock()
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("HTTP GET body error: %s\n", err.Error())
			resp.Body.Close()
			situationMutex.Unlock()
			continue
		}

		//		fmt.Printf("body: %s\n", string(body))
		err = json.Unmarshal(body, &Location)

		if err != nil {
			fmt.Printf("HTTP JSON unmarshal error: %s\n", err.Error())
		}
		resp.Body.Close()
		situationMutex.Unlock()

	}
}

func main() {
	situationMutex = &sync.Mutex{}

	rfm95w, err := goRFM95W.New()
	chkErr(err)

	go situationUpdater() // Keep the GPS location updated.

	// Wait for incoming messages on the LoRa link.
	rfm95w.Start()
	for {
		time.Sleep(1 * time.Second)
		msgs := rfm95w.FlushRXBuffer()
		if len(msgs) > 0 {
			for _, msg := range msgs {
				fmt.Printf("%f,%f,%f,%s,%d dBm,%f dB,%0.3f MHz,%d kHz,%d,%d,%d,%s\n", Location.Lat, Location.Lng, Location.Alt, Location.GPSTime, msg.RSSI, msg.SNR, float32(msg.Params.Frequency)/float32(1000000.0), msg.Params.Bandwidth/1000, msg.Params.SpreadingFactor, msg.Params.CodingRate, msg.Params.PreambleLength, msg.Buf)
			}
		}
	}
}
