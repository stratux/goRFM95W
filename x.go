package main

import (
	"./goRFM95W"
	"fmt"
	"github.com/klauspost/reedsolomon"
	"github.com/sigurn/crc8"
	"time"
)

// Data format is (shard1)(crc8_shard1)(shard2)(crc8_shard2)...(rs_shard1)(rs_shard2)...
func encode(msg []byte) []byte {
	table := crc8.MakeTable(crc8.CRC8_MAXIM)

	data := make([][]byte, 43)
	crc := make([]byte, 40)
	for i := 0; i < 43; i++ {
		data[i] = make([]byte, 5)
	}

	for i := 0; i < 40; i++ {
		for j := 0; j < 5; j++ {
			data[i][j] = msg[5*i+j]
		}
		crc[i] = byte(crc8.Checksum(data[i], table))
	}
	enc, err := reedsolomon.New(40, 3)
	if err != nil {
		panic(err)
	}
	err = enc.Encode(data)
	if err != nil {
		panic(err)
	}
	ret := make([]byte, 0)
	for i := 0; i < 43; i++ {
		ret = append(ret, data[i]...)
		if i < 40 {
			ret = append(ret, crc[i])
		}
	}
	return ret
}

func decode(msg []byte) ([]byte, error) {
	table := crc8.MakeTable(crc8.CRC8_MAXIM)

	data := make([][]byte, 43)
	crc := make([]byte, 40)
	for i := 0; i < 40; i++ {
		data[i] = make([]byte, 5)
		for j := 0; j < 5; j++ {
			data[i][j] = msg[6*i+j]
		}
		crc[i] = msg[6*i+5]
		crc_received := byte(crc8.Checksum(data[i], table))
		if crc[i] != crc_received {
			fmt.Printf("crc err, block %d\n", i)
			data[i] = nil // Delete this block for reconstruction by RS algorithm.
		}
	}

	// Fill in the RS parity shards.
	data[40] = msg[240:245]
	data[41] = msg[245:250]
	data[42] = msg[250:255]

	enc, _ := reedsolomon.New(40, 3)
	err := enc.Reconstruct(data)

	if err != nil {
		return nil, err
	}

	ret := make([]byte, 0)
	for i := 0; i < 40; i++ {
		ret = append(ret, data[i]...)
	}
	return ret, nil
}

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

	buf_encoded := encode(buf)

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
				msg_corrected, _ := decode(msg.Buf)
				fmt.Printf("Message: ")
				for i := 0; i < len(msg_corrected); i++ {
					fmt.Printf("%02x ", msg_corrected[i])
				}
				fmt.Printf("\n")
			}
		}
	}
}
