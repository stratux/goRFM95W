package goRFM95W

import (
	"errors"
	"github.com/cyoung/rpi"
	"golang.org/x/exp/io/spi"
	"time"
)

type RFM95W struct {
	SPI   *spi.Device
	mode  int
	freq  uint64 // Hz
	bw    int    // kHz
	sf    int
	cr    int
	pr    int
	txpwr int
}

const (
	RF95W_DEFAULT_FREQ  = 915000000 // Hz
	RF95W_DEFAULT_BW    = 500       // kHz
	RF95W_DEFAULT_SF    = 12
	RF95W_DEFAULT_CR    = 4
	RF95W_DEFAULT_PR    = 8
	RF85W_DEFAULT_TXPWR = 13 // dBm

	// Hardware config.
	RF95W_CS_PIN       = rpi.PIN_GPIO_10
	RF95W_DIO0_INT_PIN = rpi.PIN_GPIO_6

	SPI_WRITE_MASK = 0x80
)

func New() (*RFM95W, error) {
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
		SPI:   SPI,
		mode:  0, // FIXME.
		freq:  RF95W_DEFAULT_FREQ,
		bw:    RF95W_DEFAULT_BW,
		sf:    RF95W_DEFAULT_SF,
		cr:    RF95W_DEFAULT_CR,
		pr:    RF95_DEFAULT_PR,
		txpwr: RF85W_DEFAULT_TXPWR,
	}

	time.Sleep(100 * time.Millisecond)

	err = ret.init()

	return ret, err
}

func (r *RFM95W) SetMode(mode byte) error {
	_, err := r.SetRegister(0x01, mode)
	return err
}

/*
	Close().
	 Cleanup functions. Shut down the module and close the SPI handle.
*/

func (r *RFM95W) Close() {
	// Put the module to sleep when it is not in use.
	r.SetMode(RF95W_MODE_SLEEP)
	r.SPI.Close()
}

func (r *RFM95W) init() error {
	r.SetMode(RF95W_MODE_SLEEP | RF95W_MODE_LORA)

	time.Sleep(10 * time.Millisecond)

	mode, err := r.GetMode()
	if err != nil {
		return err
	}

	// Use the "mode" setting to check connection.
	if mode != RF95W_MODE_SLEEP|RF95W_MODE_LORA {
		return errors.New("Init failed - couldn't set mode on module.")
	}

	//TODO: WiringPi interrupts.

	// Set base addresses of the FIFO buffer in both TX and RX cases to zero.
	r.SetRegister(0x0E, 0x00)
	r.SetRegister(0x0F, 0x00)

	// Set module to STDBY mode.
	r.SetMode(RF95W_MODE_STDBY)

	r.SetBandwidth(r.bw)
	r.SetSpreadingFactor(r.sf)
	r.SetCodingRate(r.cr)

	r.SetPreambleLength(r.pr)
	r.SetFrequency(r.freq)
	r.SetTXPower(r.txpwr)

	return nil
}

//TODO
func (r *RFM95W) SetBandwidth(bw int) error {

}

//TODO
func (r *RFM95W) SetSpreadingFactor(sf int) error {

}

//TODO
func (r *RFM95W) SetCodingRate(cr int) error {

}

//TODO
func (r *RFM95W) SetPreambleLength(pr int) error {

}

//TODO
func (r *RFM95W) SetFrequency(freq uint64) error {

}

//TODO
func (r *RFM95W) SetTXPower(txpwr int) error {

}
