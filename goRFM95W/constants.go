package goRFM95W

const (
	RF95W_MODE_FSK          = 0x00
	RF95W_MODE_OOK          = 0x10
	RF95W_MODE_LORA         = 0x80
	RF95W_MODE_SLEEP        = 0x00
	RF95W_MODE_STDBY        = 0x01
	RF95W_MODE_FSTX         = 0x02
	RF95W_MODE_TX           = 0x03
	RF95W_MODE_FSRX         = 0x04
	RF95W_MODE_RXCONTINUOUS = 0x05
	RF95W_MODE_RXSINGLE     = 0x06 // LoRa specific.
	RF95W_MODE_CAD          = 0x07 // LoRa specific.

	RF95W_FREQ_STEP = 32000000.0 / 524288.0 // 32 MHz oscillator, 2^19 bits. ~61 Hz.

	RF95W_IRQ_FLAG_RXTIMEOUT         = 0x80
	RF95W_IRQ_FLAG_RXDONE            = 0x40
	RF95W_IRQ_FLAG_PAYLOADCRCERROR   = 0x20
	RF95W_IRQ_FLAG_VALIDHEADER       = 0x10
	RF95W_IRQ_FLAG_TXDONE            = 0x08
	RF95W_IRQ_FLAG_CADDONE           = 0x04
	RF95W_IRQ_FLAG_FHSSCHANGECHANNEL = 0x02
	RF95W_IRQ_FLAG_CADDETECTED       = 0x01

	MAX_TXQUEUE_PILEUP = 100000 // About 25MB of messages. Start dropping messages in the queue once reaching this.
)

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
	250000: 0x8,
	500000: 0x9,
}
