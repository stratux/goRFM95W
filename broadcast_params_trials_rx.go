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
	// Start the GPS logging functions in the background.
	situationMutex = &sync.Mutex{}

	rfm95w, err := goRFM95W.New()
	chkErr(err)

	go situationUpdater() // Keep the GPS location updated.

	rfm95w.Start()

	i := 0
	for {
		param := testParams[i]
		fmt.Printf("%d %v\n", i, param)
		err := rfm95w.SetParams(param)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
		}

		recvTimeout := time.After(150*time.Millisecond + (4 * testParamsTXTime[i]))
		checkMsgs := time.Tick(10 * time.Millisecond)
		for {
			finished := false
			select {
			case <-checkMsgs:
				msgs := rfm95w.FlushRXBuffer()
				if len(msgs) > 0 {
					for _, msg := range msgs {
						fmt.Printf("%f,%f,%f,%s,%d dBm,%f dB,%0.3f MHz,%0.3f kHz,%d,%d,%d,%s\n", Location.Lat, Location.Lng, Location.Alt, Location.GPSTime, msg.RSSI, msg.SNR, float32(msg.Params.Frequency)/float32(1000000.0), float32(msg.Params.Bandwidth)/float32(1000.0), msg.Params.SpreadingFactor, msg.Params.CodingRate, msg.Params.PreambleLength, msg.Buf)
					}
					// Got a message. Move on to the next parameter.
					i++
					if i >= len(testParams) {
						i = 0
					}
					finished = true
				}
			case <-recvTimeout:
				if i > 0 {
					// Receive timeout. Move backwards in the list (higher sensitivity).
					i--
				}
				finished = true
			}
			if finished {
				break
			}
		}
	}
}
