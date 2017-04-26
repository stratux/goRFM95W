package goRFM95W

import (
	"github.com/cyoung/rpi"
)

func (r *RFM95W) GetBytes(reg byte, len int) ([]byte, error) {
	rpi.DigitalWrite(RF95W_CS_PIN, rpi.LOW)
	defer rpi.DigitalWrite(RF95W_CS_PIN, rpi.HIGH)

	buf := make([]byte, len+1)
	err := r.SPI.Tx([]byte{byte(uint32(reg) & ^uint32(SPI_WRITE_MASK)), 0x00}, buf)
	if err != nil {
		return nil, err
	}
	return buf[1:], nil
}

func (r *RFM95W) SetBytes(reg byte, val []byte) ([]byte, error) {
	rpi.DigitalWrite(RF95W_CS_PIN, rpi.LOW)
	defer rpi.DigitalWrite(RF95W_CS_PIN, rpi.HIGH)

	outBuf := []byte{reg | SPI_WRITE_MASK}
	outBuf = append(outBuf, val...)
	inBuf := make([]byte, 1+len(val))

	err := r.SPI.Tx(outBuf, inBuf)
	if err != nil {
		return nil, err
	}

	return inBuf[1:], nil
}

func (r *RFM95W) GetRegister(reg byte) (byte, error) {
	var ret byte
	x, err := r.GetBytes(reg, 1)
	if x != nil {
		ret = x[0]
	}
	return ret, err
}

func (r *RFM95W) SetRegister(reg, val byte) (byte, error) {
	var ret byte
	x, err := r.SetBytes(reg, []byte{val})
	if x != nil {
		ret = x[0]
	}
	return ret, err
}
