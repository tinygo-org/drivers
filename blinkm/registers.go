package blinkm

// Constants/addresses used for BlinkM.

// The I2C address which this device listens to.
const Address = 0x09

// Registers, which in the case of the BlinkM are actually commands.
const (
	TO_RGB            = 0x6e
	FADE_TO_RGB       = 0x63
	FADE_TO_HSB       = 0x68
	FADE_TO_RND_RGB   = 0x43
	FADE_TO_RND_HSB   = 0x48
	PLAY_LIGHT_SCRIPT = 0x70
	STOP_SCRIPT       = 0x6f
	SET_FADE          = 0x66
	SET_TIME          = 0x74
	GET_RGB           = 0x67
	GET_ADDRESS       = 0x61
	SET_ADDRESS       = 0x41
	GET_FIRMWARE      = 0x5a
)
