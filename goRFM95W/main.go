package goRFM95W

import (
	"github.com/cyoung/rpi"
	"golang.org/x/exp/io/spi"
	"time"
)

type RFM95W struct {
	SPI  *spi.Device
	mode int
	freq float64 // MHz
	bw   int     // kHz
	sf   int
	cr   int
}

const (
	RF95W_DEFAULT_FREQ = 915 // MHz
	RF95W_DEFAULT_BW   = 500 // kHz
	RF95W_DEFAULT_SF   = 12
	RF95W_DEFAULT_CR   = 4

	// Hardware config.
	RF95W_CS_PIN       = rpi.PIN_GPIO_10
	RF95W_DIO0_INT_PIN = rpi.PIN_GPIO_6

	SPI_WRITE_MASK = 0x80
)

func NewRFM95W() (*RFM95W, error) {
	// Initialize GPIO library.
	rpi.WiringPiSetup()

	// Set up the CS and interrupt (DIO0) pins.
	rpi.PinMode(RF95W_CS_PIN, rpi.OUTPUT)
	rpi.PinMode(RF95W_DIO0_INT_PIN, rpi.INPUT)

	rpi.DigitalWrite(RF95W_CS_PIN, rpi.HIGH)

	spiDev := &spi.Devfs{
		Dev:      "/dev/spidev0.0",
		Mode:     spi.Mode0,
		MaxSpeed: 8000000,
	}

	SPI, err := spi.Open(spiDev)
	if err != nil {
		return nil, err
	}

	SPI.SetBitOrder(spi.MSBFirst)
	SPI.SetCSChange(false)

	ret := &RFM95W{
		SPI:  SPI,
		mode: 0, // FIXME.
		freq: RF95W_DEFAULT_FREQ,
		bw:   RF95W_DEFAULT_BW,
		sf:   RF95W_DEFAULT_SF,
		cr:   RF95W_DEFAULT_CR,
	}

	time.Sleep(100 * time.Millisecond)

	return ret, nil
}

func (r *RFM95W) Close() {
	r.SPI.Close()
}
