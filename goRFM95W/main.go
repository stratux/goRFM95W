package goRFM95W

import (
	"errors"
	"fmt"
	"github.com/cyoung/rpi"
	"golang.org/x/exp/io/spi"
	"time"
)

type RFM95W_Message struct {
	Buf      []byte
	RSSI     int
	SNR      int
	Received time.Time
}

type RFM95W struct {
	SPI           *spi.Device
	mode          int
	freq          uint64 // Hz
	bw            int    // Hz
	sf            int
	cr            int
	pr            int
	interruptChan chan int
	RecvBuf       []RFM95W_Message // This is constantly being filled up as messages are received.
	currentMode   byte
}

const (
	RF95W_DEFAULT_FREQ = 915000000 // Hz
	RF95W_DEFAULT_BW   = 500000    // Hz
	RF95W_DEFAULT_SF   = 12
	RF95W_DEFAULT_CR   = 4
	RF95W_DEFAULT_PR   = 8

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
	if err != nil {
		r.currentMOde = mode
	}
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

	// Set up the WiringPi interrupt for DIO0.
	r.interruptChan = WiringPiISR(RF95W_DIO0_INT_PIN, rpi.INT_EDGE_RISING)

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
	r.SetTXPower()

	return nil
}

/*
	SetBandwidth().
	 Sets the total bandwidth to use in the transmission.
*/

// bandwidth (Hz) -> setting.
var RFM95W_Bandwidths = map[int]byte{
	7800:   0x0,
	10400:  0x1,
	15600:  0x2,
	20800:  0x3,
	31250:  0x4,
	41700:  0x5,
	62500:  0x6,
	125000: 0x7,
	500000: 0x8,
}

func (r *RFM95W) SetBandwidth(bw int) error {
	if b, ok := RFM95W_Bandwidths[bw]; !ok {
		return errors.New("Invalid bandwidth requested.")
	}
	// Get initial value.
	val, err := r.GetRegister(0x1D)
	if err != nil {
		return err
	}
	// Set only the bandwidth portion.
	new_val := (val & 0x0F) | (b << 4)
	fmt.Printf("SetBandwidth(): %02x -> %02x\n", val, new_val)
	_, err = r.SetRegister(0x1D, new_val)
	return err
}

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
	val, err := r.GetRegister(0x1D)
	if err != nil {
		return err
	}
	// Set only the coding rate portion.
	new_val := (val & 0xF1) | (b << 1)
	fmt.Printf("SetCodingRate(): %02x -> %02x\n", val, new_val)
	_, err := r.SetRegister(0x1D, new_val)
	return err
}

/*
	SetExplicitHeaderMode().
	 True or false - include explicit header.
	 Currently always setting "false" on init since no other header handling is implemented.
*/

func (r *RFM95W) SetExplicitHeaderMode(wantHeader bool) error {
	// Get initial value.
	val, err := r.GetRegister(0x1D)
	if err != nil {
		return err
	}
	var b byte
	if !wantHeader {
		b = 0x1
	}
	// Set only the header portion.
	new_val := (val & 0xFE) | b
	fmt.Printf("SetExplicitHeaderMode(): %02x -> %02x\n", val, new_val)
	_, err := r.SetRegister(0x1D, new_val)
	return err
}

/*
	SetSpreadingFactor().
	 Sets the spreading factor. Valid values are 6, 7, 8, 9, 10, 11, 12.
*/

func (r *RFM95W) SetSpreadingFactor(sf int) error {
	if sf < 6 || sf > 12 {
		return errors.New("Invalid spreading factor requested.")
	}
	b := byte(sf)
	// Get initial value.
	val, err := r.GetRegister(0x1E)
	if err != nil {
		return err
	}
	// Set only the spreading factor portion.
	new_val := (val & 0x0F) | (b << 4)
	fmt.Printf("SetSpreadingFactor(): %02x -> %02x\n", val, new_val)
	_, err := r.SetRegister(0x1E, new_val)
	return err
}

/*
	SetPreambleLength().
	 Sets the preamble length, from 6-65535.
	 Default value is 8.
*/
func (r *RFM95W) SetPreambleLength(pr int) error {
	if pr < 6 {
		return errors.New("Invalid preamble length requested.")
	}
	r.SetRegister(0x20, byte(pr>>8))
	_, err := r.SetRegister(0x21, byte(pr&0xFF))
	return err
}

func (r *RFM95W) SetFrequency(freq uint64) error {
	steps := uint32(float64(freq) / RF95W_FREQ_STEP)
	r.SetRegister(0x06, steps>>16)
	r.SetRegister(0x07, (steps>>8)&0xFF)
	r.SetRegister(0x08, (steps & 0xFF))
}

/*
	SetTXPower().
//FIXME:	 Always run at 17 dBm for now.
*/
func (r *RFM95W) SetTXPower() error {
	_, err := r.SetRegister(0x09, 0x8F)
	return err
}

/*
	Send().
	 Sends a single message. Stops the receive thread and switches to TX mode until the message is sent.
*/

func (r *RFM95W) Send(msg []byte) error {
	if len(msg) > 255 {
		return errors.New("Message too long.")
	}

	//FIXME: Stop the receive thread.

	r.SetMode(RF95W_MODE_STDBY)

	// Set the FIFO address pointer to the start.
	_, err := r.SetRegister(0x0D, 0x00)
	if err != nil {
		return err
	}

	// Write the message into the FIFO buffer.
	_, err = r.SetBytes(0x00, msg)
	if err != nil {
		return err
	}

	// Change DIOx register mapping so that DIO0 interrupts when transmission has finished.
	_, err = r.SetRegister(0x40, 0x40)
	if err != nil {
		return err
	}

	// Begin transmitting.
	err = r.SetMode(RF95W_MODE_TX)
	return err
}

//TODO
func (r *RFM95W) queueHandler() {

}
