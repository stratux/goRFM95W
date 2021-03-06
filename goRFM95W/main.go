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

type RFM95W_Params struct {
	TransmitMode    int
	Frequency       uint64 // Hz.
	Bandwidth       int    // Hz.
	SpreadingFactor int    // LoRa specific.
	CodingRate      int    // LoRa specific.
	PreambleLength  int    // LoRa specific.
	DataRate        int    // FSK specific.
}

type RFM95W_Message struct {
	Buf      []byte
	RSSI     int     // dBm
	SNR      float64 // dB
	Received time.Time
	Params   RFM95W_Params
}

type RFM95W struct {
	Debug         bool // Print debug messages
	SPI           *spi.Device
	mode          int
	settings      RFM95W_Params
	interruptChan chan int
	mu_Recv       *sync.Mutex
	RecvBuf       []RFM95W_Message // This is constantly being filled up as messages are received.
	txQueue       chan []byte
	mu_Send       *sync.Mutex
	currentMode   byte
	stopQueue     chan int
	// Temp variables for stats.
	txStart    time.Time
	LastTXTime time.Duration
}

// Default settings.
const (
	RF95W_DEFAULT_FREQ = 915000000 // Hz
	RF95W_DEFAULT_BW   = 500000    // Hz
	RF95W_DEFAULT_SF   = 11
	RF95W_DEFAULT_CR   = 5
	RF95W_DEFAULT_PR   = 8

	// Hardware config.
	RF95W_CS_PIN       = rpi.PIN_CE0
	RF95W_DIO0_INT_PIN = rpi.PIN_GPIO_6
	RF95W_ACT_PIN      = rpi.PIN_CE1

	SPI_WRITE_MASK = 0x80
)

func New(params *RFM95W_Params) (*RFM95W, error) {
	if params == nil {
		// Default parameters.
		params = &RFM95W_Params{
			TransmitMode:    RF95W_MODE_LORA,
			Frequency:       RF95W_DEFAULT_FREQ,
			Bandwidth:       RF95W_DEFAULT_BW,
			SpreadingFactor: RF95W_DEFAULT_SF,
			CodingRate:      RF95W_DEFAULT_CR,
			PreambleLength:  RF95W_DEFAULT_PR,
		}
	}
	// Initialize GPIO library.
	rpi.WiringPiSetup()

	// Set up the CS and interrupt (DIO0) pins.
	rpi.PinMode(RF95W_CS_PIN, rpi.OUTPUT)      // Chip Select.
	rpi.PinMode(RF95W_DIO0_INT_PIN, rpi.INPUT) // DIO0 interrupt.
	rpi.PinMode(RF95W_ACT_PIN, rpi.OUTPUT)     // ACT LED.

	rpi.DigitalWrite(RF95W_CS_PIN, rpi.HIGH)
	rpi.DigitalWrite(RF95W_ACT_PIN, rpi.LOW)

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
		SPI:      SPI,
		mode:     0, // FIXME.
		settings: *params,
	}

	// Variables that need initializing.
	ret.txQueue = make(chan []byte, 1024)
	ret.stopQueue = make(chan int)
	ret.mu_Recv = &sync.Mutex{}
	ret.mu_Send = &sync.Mutex{}

	time.Sleep(100 * time.Millisecond)

	err = ret.init()

	return ret, err
}

func (r *RFM95W) SetMode(mode byte) error {
	_, err := r.SetRegister(0x01, mode) // RegOpMode.
	if err == nil {
		r.currentMode = mode
	}
	return err
}

func (r *RFM95W) GetMode() (byte, error) {
	ret, err := r.GetRegister(0x01) // RegOpMode.
	if err == nil {
		r.currentMode = ret
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

func (r *RFM95W) setParams(param RFM95W_Params) {
	r.SetBandwidth(param.Bandwidth)
	r.SetSpreadingFactor(param.SpreadingFactor)
	r.SetCodingRate(param.CodingRate)
	r.SetPreambleLength(param.PreambleLength)
	r.SetFrequency(param.Frequency)
}

func (r *RFM95W) SetParams(param RFM95W_Params) error {
	if r.currentMode == RF95W_MODE_TX {
		return errors.New("SetParams(): Not ready.")
	}

	r.SetMode(RF95W_MODE_STDBY)

	r.settings = param
	err := r.init()

	if err != nil {
		return err
	}

	r.setRXMode()

	return nil
}

func (r *RFM95W) init() error {
	i := 0
	for i = 0; i < 10; i++ {
		// Retry setting the mode 10 times.
		r.SetMode(RF95W_MODE_SLEEP | RF95W_MODE_LORA)

		time.Sleep(10 * time.Millisecond)

		mode, err := r.GetMode()
		if err != nil {
			return err
		}

		// Use the "mode" setting to check connection.
		if mode == RF95W_MODE_SLEEP|RF95W_MODE_LORA {
			break
		}
	}
	if i == 10 { // 10 retries was not enough - some issue.
		return errors.New("Init failed - couldn't set mode on module.")
	}

	// Set up the WiringPi interrupt for DIO0, if it is not yet set up.
	if r.interruptChan == nil {
		r.interruptChan = rpi.WiringPiISR(RF95W_DIO0_INT_PIN, rpi.INT_EDGE_RISING)
	}

	// Set base addresses of the FIFO buffer in both TX and RX cases to zero.
	r.SetRegister(0x0E, 0x00) // RegFifoTxBaseAddr.
	r.SetRegister(0x0F, 0x00) // RegFifoRxBaseAddr.

	// Set module to STDBY mode.
	r.SetMode(RF95W_MODE_STDBY)

	r.setParams(r.settings)

	r.setLNASettings()

	r.SetTXPower()

	return nil
}

/*
	setLNASettings().
	 Sets the LNA gain and boost values.
*/
func (r *RFM95W) setLNASettings() {
	// G1 = maximum gain. LnaBoostHf on.
	r.SetRegister(0x0C, 0x23) // RegLna.
	//FIXME: See RegModemConfig3. AgcAutoOn.
}

/*
	SetBandwidth().
	 Sets the total bandwidth to use in the transmission.
*/

func (r *RFM95W) SetBandwidth(bw int) error {
	b, ok := RFM95W_Bandwidths[bw]
	if !ok {
		return errors.New("Invalid bandwidth requested.")
	}
	// Get initial value.
	val, err := r.GetRegister(0x1D) // RegModemConfig1.
	if err != nil {
		return err
	}
	// Set only the bandwidth portion.
	new_val := (val & 0x0F) | (b << 4)
	if r.Debug {
		fmt.Printf("SetBandwidth(): %02x -> %02x\n", val, new_val)
	}
	_, err = r.SetRegister(0x1D, new_val) // RegModemConfig1.
	if err != nil {
		r.settings.Bandwidth = bw
	}
	return err
}

/*
	SetExplicitHeaderMode().
	 True or false - include explicit header.
	 Currently always setting "false" on init since no other header handling is implemented.
*/

func (r *RFM95W) SetExplicitHeaderMode(wantHeader bool) error {
	// Get initial value.
	val, err := r.GetRegister(0x1D) // RegModemConfig1.
	if err != nil {
		return err
	}
	var b byte
	if !wantHeader {
		b = 0x1
	}
	// Set only the header portion.
	new_val := (val & 0xFE) | b
	if r.Debug {
		fmt.Printf("SetExplicitHeaderMode(): %02x -> %02x\n", val, new_val)
	}
	_, err = r.SetRegister(0x1D, new_val) // RegModemConfig1.
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
	r.SetRegister(0x20, byte(pr>>8))             // RegPreambleMsb.
	_, err := r.SetRegister(0x21, byte(pr&0xFF)) // RegPreambleLsb.
	if err != nil {
		r.settings.PreambleLength = pr
	}
	return err
}

func (r *RFM95W) SetFrequency(freq uint64) error {
	steps := uint32(float64(freq) / RF95W_FREQ_STEP)
	r.SetRegister(0x06, byte(steps>>16))            // RegFrMsb.
	r.SetRegister(0x07, byte((steps>>8)&0xFF))      // RegFrMid.
	_, err := r.SetRegister(0x08, byte(steps&0xFF)) // RegFrLsb.
	if err != nil {
		r.settings.Frequency = freq
	}
	return err
}

/*
	SetTXPower().
//FIXME:	 Always run at 17 dBm for now.
*/
func (r *RFM95W) SetTXPower() error {
	_, err := r.SetRegister(0x09, 0x8F) // RegPaConfig.
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
	return nil
}

/*
	SendSync().
	 Sends the message. Waits for the queue to be emptied before returning.
*/

func (r *RFM95W) SendSync(msg []byte) error {
	err := r.sendMessage(msg)
	if err == nil {
		for r.currentMode == RF95W_MODE_TX {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return err
}

func (r *RFM95W) sendMessage(msg []byte) error {
	r.mu_Send.Lock()
	defer r.mu_Send.Unlock()

	r.SetMode(RF95W_MODE_STDBY)

	rpi.DigitalWrite(RF95W_ACT_PIN, rpi.HIGH) // Turn on ACT LED.

	// Set the FIFO address pointer to the start.
	_, err := r.SetRegister(0x0D, 0x00) // RegFifoAddrPtr.
	if err != nil {
		return err
	}

	// Write the message into the FIFO buffer.
	_, err = r.SetBytes(0x00, msg) // RegFifo.
	if err != nil {
		return err
	}

	// Set the message payload length register.
	_, err = r.SetRegister(0x22, byte(len(msg))) // PayloadLength.
	if err != nil {
		return err
	}

	// Change DIOx interrupt mapping so that DIO0 interrupts on TxDone.
	_, err = r.SetRegister(0x40, 0x40) // RegDioMapping1.
	if err != nil {
		return err
	}

	// Begin transmitting.
	r.txStart = time.Now()
	err = r.SetMode(RF95W_MODE_TX)
	return err
}

func (r *RFM95W) setRXMode() error {
	err := r.SetMode(RF95W_MODE_RXCONTINUOUS)
	if err != nil {
		return err
	}

	// Change DIOx interrupt mapping so that DIO0 interrupts on RxDone.
	_, err = r.SetRegister(0x40, 0x00) // RegDioMapping1.
	if err != nil {
		return err
	}
	return nil
}

/*
	queueHandler().
	 Receives TX messages and coordinates transmissions between RX. TX takes priority, and the default mode of opreation is
	 "RXCONTINUOUS".
//FIXME: Check 0x18 RegModemStat/ModemStatus to see if "RX on-going" before switching to TX.
*/

func (r *RFM95W) queueHandler() {
	//FIXME: Assuming that we're ready to start sending/receiving once this goroutine is started.
	err := r.setRXMode()
	if err != nil {
		if r.Debug {
			fmt.Printf("queueHandler() can't set receive mode: %s\n", err.Error())
		}
		return
	}

	txWaiting := make([][]byte, 0)
	for {
		select {
		case <-r.interruptChan:
			// Get the IRQ flags.
			irqFlags, _ := r.GetRegister(0x12) // RegIrqFlags.
			if r.Debug {
				fmt.Printf("queueHandler() interrupt received, currentMode=%02x, irqFlags=%02x\n", r.currentMode, irqFlags)
			}
			switch r.currentMode {
			case RF95W_MODE_TX:
				if irqFlags&RF95W_IRQ_FLAG_TXDONE != 0 {
					// TX finished.
					txEnd := time.Now()
					r.LastTXTime = txEnd.Sub(r.txStart)
					if r.Debug {
						fmt.Printf("queueHandler() transmit finished, t=%dms.\n", r.LastTXTime/time.Millisecond)
					}
					// Are there more messages that we need to send? Always empty the queue before starting to receive.
					if len(txWaiting) > 0 {
						if r.Debug {
							fmt.Printf("queuehandler() starting new transmission.\n")
						}
						// Switch to transmit mode (again).
						err := r.sendMessage(txWaiting[0])
						if err != nil {
							fmt.Printf("queueHandler() send message error: %s\n", err.Error())
						} else {
							txWaiting = txWaiting[1:] // Message was buffered to the radio successfully.
						}
					} else {
						// No more messages waiting to transmit, go back to receive mode.
						if r.Debug {
							fmt.Printf("queueHandler() finished sending all TX messages, switching back to RX mode.\n")
						}
						rpi.DigitalWrite(RF95W_ACT_PIN, rpi.LOW) // Turn off ACT LED.
						r.setRXMode()
					}
				}
			case RF95W_MODE_RXCONTINUOUS:
				if irqFlags&RF95W_IRQ_FLAG_RXTIMEOUT != 0 {
					// Timeout. Do nothing, since we're receiving in continuous mode.
				} else if irqFlags&RF95W_IRQ_FLAG_PAYLOADCRCERROR != 0 {
					fmt.Printf("queueHandler() received packet with CRC error. discarding.\n")
				} else if irqFlags&RF95W_IRQ_FLAG_RXDONE != 0 {
					if r.Debug {
						fmt.Printf("queueHandler() received RXDONE.\n")
					}
					// Get the total length of the packet.
					msgLen, err := r.GetRegister(0x13) // FifoRxNbBytes.
					if err != nil {
						fmt.Printf("queueHandler() fatal error receiving packet, can't get length: %s\n", err.Error())
						continue
					}
					// Get the start address in the FIFO queue.
					fifoPtr, err := r.GetRegister(0x10) // RegFifoRxCurrentAddr.
					if err != nil {
						fmt.Printf("queueHandler() fatal error receiving packet, can't get start pointer address: %s\n", err.Error())
						continue
					}
					// Set the read address to the start of the message in the FIFO queue.
					_, err = r.SetRegister(0x0D, fifoPtr) // RegFifoAddrPtr.
					if err != nil {
						fmt.Printf("queueHandler() fatal error receiving packet, can't set FIFO pointer: %s\n", err.Error())
						continue
					}
					// Read the data.
					msgBuf, err := r.GetBytes(0x00, int(msgLen)) // RegFifo.
					if err != nil {
						fmt.Printf("queueHandler() fatal error receiving packet, can't read FIFO buffer: %s\n", err.Error())
						continue
					}
					// Get some extra stats - SNR, RSSI, etc.
					snrByte, _ := r.GetRegister(0x19)  // RegPktSnrValue.
					rssiByte, _ := r.GetRegister(0x1A) // RegPktRssiValue.
					//FIXME: Converting snr should be easier.
					rdr := bytes.NewReader([]byte{snrByte})
					var snr int8
					binary.Read(rdr, binary.LittleEndian, &snr)
					var newMessage RFM95W_Message
					newMessage.SNR = float64(snr) / 4.0
					newMessage.RSSI = int(rssiByte) - 137
					newMessage.Buf = msgBuf
					newMessage.Received = time.Now()
					newMessage.Params = r.settings
					r.mu_Recv.Lock()
					r.RecvBuf = append(r.RecvBuf, newMessage)
					r.mu_Recv.Unlock()

				}
			}
			// Clear the IRQ flags.
			r.SetRegister(0x12, 0xFF) // RegIrqFlags.
		case msg := <-r.txQueue:
			txWaiting = append(txWaiting, msg) // txWaiting is a FIFO queue.
			if len(txWaiting) > MAX_TXQUEUE_PILEUP {
				// Too many messages are in the queue. Start dropping the oldest.
				fmt.Printf("WARNING: queueHandler() dropping oldest messages, %d in queue.\n", len(txWaiting))
				txWaiting = txWaiting[len(txWaiting)-MAX_TXQUEUE_PILEUP:]
			}
			if r.currentMode != RF95W_MODE_TX { // If we're currently in TX mode, let the current transmission finish.
				if r.Debug {
					fmt.Printf("queuehandler() starting new transmission.\n")
				}
				// Switch to transmit mode.
				err := r.sendMessage(txWaiting[0])
				if err != nil {
					fmt.Printf("queueHandler() send message error: %s\n", err.Error())
				} else {
					txWaiting = txWaiting[1:] // Message was buffered to the radio successfully.
				}
			}
		case <-r.stopQueue:
			if r.Debug {
				fmt.Printf("queueHandler() received shutdown.\n")
			}
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
	if r.Debug {
		fmt.Printf("Stopping queue thread...\n")
	}
	r.stopQueue <- 1
}

/*
	FlushRXBuffer().
	 Get all data in RecvBuf and clear it.
*/

func (r *RFM95W) FlushRXBuffer() []RFM95W_Message {
	r.mu_Recv.Lock()
	ret := r.RecvBuf
	r.RecvBuf = make([]RFM95W_Message, 0)
	r.mu_Recv.Unlock()
	return ret
}
