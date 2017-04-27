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

type RFM95W_Message struct {
	Buf      []byte
	RSSI     int     // dBm
	SNR      float64 // dB
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
	mu_Recv       *sync.Mutex
	RecvBuf       []RFM95W_Message // This is constantly being filled up as messages are received.
	txQueue       chan []byte
	currentMode   byte
	stopQueue     chan int
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

	// Variables that need initializing.
	ret.txQueue = make(chan []byte, 1024)
	ret.stopQueue = make(chan int)
	ret.mu_Recv = &sync.Mutex{}

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

func (r *RFM95W) GetMode() (byte, error) {
	ret, err := r.GetRegister(0x01)
	if err != nil {
		currentMode = ret
	}
	return ret, err
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
	 Sends a single message. Sends the message to the txQueue channel for processing by queueHandler.
*/

func (r *RFM95W) Send(msg []byte) error {
	if len(msg) > 255 {
		return errors.New("Message too long.")
	}

	r.txQueue <- msg
}

func (r *RFM95W) sendMessage(msg []byte) error {
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

	// Change DIOx interrupt mapping so that DIO0 interrupts on TxDone.
	_, err = r.SetRegister(0x40, 0x40)
	if err != nil {
		return err
	}

	// Begin transmitting.
	err = r.SetMode(RF95W_MODE_TX)
	return err
}

//TODO: Put the receiver into STDBY when SetFrequency, etc are called.
func (r *RFM95W) queueHandler() {
	//FIXME: Assuming that we're ready to start sending/receiving once this goroutine is started.
	err := r.SetMode(RF95W_MODE_RXCONTINUOUS)
	if err != nil {
		fmt.Printf("queueHandler() can't set receive mode: %s\n", err.Error())
		return
	}

	//Change DIOx interrupt mapping so that DIO0 interrupts on RxDone.
	_, err = r.SetRegister(0x40, 0x00)
	if err != nil {
		fmt.Printf("queueHandler() can't set up interrupt: %s\n", err.Error())
		return
	}

	txWaiting = make([][]byte, 0)
	for {
		select {
		case <-r.interruptChan:
			// Get the IRQ flags.
			irqFlags := r.GetRegister(0x12)
			fmt.Printf("queueHandler() interrupt received, currentMode=%02x, irqFlags=%02x\n", currentMOde, irqFlags)
			switch r.currentMode {
			case RF95W_MODE_TX:
				if irqFlags&RF95W_IRQ_FLAG_TXDONE != 0 {
					// TX finished.
					fmt.Printf("queueHandler() transmit finished.\n")
					// Are there more messages that we need to send? Always empty the queue before starting to receive.
					if len(txWaiting) > 0 {
						fmt.Printf("queuehandler() starting new transmission.\n")
						// Switch to transmit mode (again).
						err := r.sendMessage(txWaiting[0])
						if err != nil {
							fmt.Printf("queueHandler() send message error: %s\n", err.Error())
						} else {
							txWaiting = txWaiting[1:] // Message was buffered to the radio successfully.
						}
					} else {
						// No more messages waiting to transmit, go back to receive mode.
						r.SetMode(RF95W_MODE_RXCONTINUOUS)
					}
				}
			case RF95W_MODE_RXCONTINUOUS:
				if irqFlags&RF95W_IRQ_FLAG_RXTIMEOUT != 0 {
					// Timeout. Do nothing, since we're receiving in continuous mode.
				} else if irqFlags&RF95W_IRQ_FLAG_PAYLOADCRCERROR != 0 {
					fmt.Printf("queueHandler() received packet with CRC error. discarding.\n")
				} else if irqFlags&RF95W_IRQ_FLAG_RXDONE != 0 {
					fmt.Printf("queueHandler() received RXDONE.\n")
					// Get the total length of the packet.
					len, err := r.GetRegister(0x13)
					if err != nil {
						fmt.Printf("queueHandler() fatal error receiving packet, can't get length: %s\n", err.Error())
						continue
					}
					// Get the start address in the FIFO queue.
					fifoPtr, err := r.GetRegister(0x10)
					if err != nil {
						fmt.Printf("queueHandler() fatal error receiving packet, can't get start pointer address: %s\n", err.Error())
						continue
					}
					// Set the read address to the start of the message in the FIFO queue.
					err = r.SetRegister(0x0D, fifoPtr)
					if err != nil {
						fmt.Printf("queueHandler() fatal error receiving packet, can't set FIFO pointer: %s\n", err.Error())
						continue
					}
					// Read the data.
					msgBuf, err := r.GetBytes(0x00, int(len))
					if err != nil {
						fmt.Printf("queueHandler() fatal error receiving packet, can't read FIFO buffer: %s\n", err.Error())
						continue
					}
					// Get some extra stats - SNR, RSSI, etc.
					snrByte, _ := r.GetRegister(0x19)
					rssiByte, _ := r.GetRegister(0x1A)
					//FIXME: Converting snr should be easier.
					rdr := bytes.NewReader([]byte{snrByte})
					var snr int8
					binary.Read(buf, binary.LittleEndian, &snr)
					var newMessage RFM95W_Message
					newMessage.SNR = float64(snr) / 4.0
					newMessage.RSSI = int(rssiByte) - 137
					newMessage.Buf = msgBuf
					newMessage.Received = time.Now()
					fmt.Printf("Message: %v\n", newMessage)
					r.mu_Recv.Lock()
					r.RecvBuf = append(r.RecvBuf, newMessage)
					r.mu_Recv.Unlock()
				}
			}
			// Clear the IRQ flags.
			r.SetRegister(0x12, 0xFF)
		case msg := <-r.txQueue:
			txWaiting = append(txWaiting, msg) // txWaiting is a FIFO queue.
			if len(txWaiting) > MAX_TXQUEUE_PILEUP {
				// Too many messages are in the queue. Start dropping the oldest.
				fmt.Printf("WARNING: queueHandler() dropping oldest messages, %d in queue.\n", len(txWaiting))
				txWaiting = txWaiting[len(txWaiting)-MAX_TXQUEUE_PILEUP:]
			}
			if r.currentMode != RF95W_MODE_TX { // If we're currently in TX mode, let the current transmission finish.
				fmt.Printf("queuehandler() starting new transmission.\n")
				// Switch to transmit mode.
				err := r.sendMessage(txWaiting[0])
				if err != nil {
					fmt.Printf("queueHandler() send message error: %s\n", err.Error())
				} else {
					txWaiting = txWaiting[1:] // Message was buffered to the radio successfully.
				}
			}
		case <-r.stopQueue:
			fmt.Printf("queueHandler() received shutdown.\n")
			r.SetMode(RF95W_MODE_STDBY)
			return
		}
	}
}

/*
	Start().
	 Starts the queue handler (TX on request and continuous RX).
	 This is called when all of the parameters and settings have been set.
*/
func (r *RFM95W) Start() {
	go r.queueHandler()
}

/*
	Stop().
	 Stops the queue handler.
	 This is called when we want to change settings.
*/
func (r *RFM95W) Stop() {
	fmt.Printf("Stopping queue thread...\n")
	r.stopQueue <- 1
}
