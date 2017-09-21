package goRFM95W

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cyoung/rpi"
	"golang.org/x/exp/io/spi"
	"sync"
	"time"
)

/*
	SetCodingRate().
	 Sets the coding rate. Valid values are 5 (4/5), 6 (4/6), 7 (4/7), 8 (4/8).
*/

func (r *RFM95W) SetCodingRate(cr int) error {
	if cr < 5 || cr > 8 {
		return errors.New("Invalid coding rate requested.")
	}
	b := byte(cr - 4) // 5 = 0x1, 6 = 0x2, 7 = 0x3, 8 = 0x4
	// Get initial value.
	val, err := r.GetRegister(0x1D) // RegModemConfig1.
	if err != nil {
		return err
	}
	// Set only the coding rate portion.
	new_val := (val & 0xF1) | (b << 1)
	if r.Debug {
		fmt.Printf("SetCodingRate(): %02x -> %02x\n", val, new_val)
	}
	_, err = r.SetRegister(0x1D, new_val) // RegModemConfig1.
	if err != nil {
		r.settings.CodingRate = cr
	}
	return err
}

/*
	SetSpreadingFactor().
	 Sets the spreading factor. Valid values are 6, 7, 8, 9, 10, 11, 12.
*/

func (r *RFM95W) SetSpreadingFactor(sf int) error {
	if sf == 6 {
		panic("SF=6 not implemented.")
	}
	if sf < 6 || sf > 12 {
		return errors.New("Invalid spreading factor requested.")
	}
	b := byte(sf)
	// Get initial value.
	val, err := r.GetRegister(0x1E) // RegModemConfig2.
	if err != nil {
		return err
	}
	// Set only the spreading factor portion.
	new_val := (val & 0x0F) | (b << 4)
	if r.Debug {
		fmt.Printf("SetSpreadingFactor(): %02x -> %02x\n", val, new_val)
	}
	_, err = r.SetRegister(0x1E, new_val) // RegModemConfig2.
	if err != nil {
		r.settings.SpreadingFactor = sf
	}
	return err
}
