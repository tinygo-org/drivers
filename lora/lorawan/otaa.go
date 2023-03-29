package lorawan

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
)

// Otaa is used to store Over The Air Activation data of a LoRaWAN session
type Otaa struct {
	DevEUI   [8]uint8
	AppEUI   [8]uint8
	AppKey   [16]uint8
	devNonce [2]uint8
	appNonce [3]uint8
	NetID    [3]uint8
	buf      []uint8
}

// Initialize DevNonce
func (o *Otaa) Init() {
	o.buf = make([]uint8, 0)
	o.generateDevNonce()
}

func (o *Otaa) generateDevNonce() {
	// TODO: handle error
	rnd, _ := GetRand16()
	o.devNonce[0] = rnd[0]
	o.devNonce[1] = rnd[1]
}

func (o *Otaa) incrementDevNonce() {
	nonce := uint16(o.devNonce[1])<<8 | uint16(o.devNonce[0]) + 1
	o.devNonce[0] = uint8(nonce)
	o.devNonce[1] = uint8((nonce >> 8))
}

// Set configures the Otaa AppEUI, DevEUI, AppKey for the device
func (o *Otaa) Set(appEUI []uint8, devEUI []uint8, appKey []uint8) {
	o.SetAppEUI(appEUI)
	o.SetDevEUI(devEUI)
	o.SetAppKey(appKey)
}

// SetAppEUI configures the Otaa AppEUI
func (o *Otaa) SetAppEUI(appEUI []uint8) error {
	if len(appEUI) != 8 {
		return ErrInvalidEuiLength
	}

	copy(o.AppEUI[:], appEUI)

	return nil
}

func (o *Otaa) GetAppEUI() string {
	return hex.EncodeToString(o.AppEUI[:])
}

// SetDevEUI configures the Otaa DevEUI
func (o *Otaa) SetDevEUI(devEUI []uint8) error {
	if len(devEUI) != 8 {
		return ErrInvalidEuiLength
	}

	copy(o.DevEUI[:], devEUI)

	return nil
}

func (o *Otaa) GetDevEUI() string {
	return hex.EncodeToString(o.DevEUI[:])
}

// SetAppKey configures the Otaa AppKey
func (o *Otaa) SetAppKey(appKey []uint8) error {
	if len(appKey) != 16 {
		return ErrInvalidAppKeyLength
	}

	copy(o.AppKey[:], appKey)

	return nil
}

func (o *Otaa) GetAppKey() string {
	return hex.EncodeToString(o.AppKey[:])
}

func (o *Otaa) GetNetID() string {
	return hex.EncodeToString(o.NetID[:])
}

func (o *Otaa) SetNetID(netID []uint8) error {
	if len(netID) != 3 {
		return ErrInvalidNetIDLength
	}

	copy(o.NetID[:], netID)

	return nil
}

// GenerateJoinRequest Generates a LoraWAN Join request
func (o *Otaa) GenerateJoinRequest() ([]uint8, error) {
	o.incrementDevNonce()

	// TODO: Add checks
	o.buf = o.buf[:0]
	o.buf = append(o.buf, 0x00)
	o.buf = append(o.buf, reverseBytes(o.AppEUI[:])...)
	o.buf = append(o.buf, reverseBytes(o.DevEUI[:])...)
	o.buf = append(o.buf, o.devNonce[:]...)
	mic := genPayloadMIC(o.buf, o.AppKey)
	o.buf = append(o.buf, mic[:]...)

	return o.buf, nil
}

// DecodeJoinAccept Decodes a Lora Join Accept packet
func (o *Otaa) DecodeJoinAccept(phyPload []uint8, s *Session) error {
	if len(phyPload) < 12 {
		return ErrInvalidPacketLength
	}
	data := phyPload[1:] // Remove trailing 0x20

	// Prepare AES Cipher
	block, err := aes.NewCipher(o.AppKey[:])
	if err != nil {
		return err
	}
	buf := make([]byte, len(data))
	for k := 0; k < len(data)/aes.BlockSize; k++ {
		block.Encrypt(buf[k*aes.BlockSize:], data[k*aes.BlockSize:])
	}

	copy(o.appNonce[:], buf[0:3])
	copy(o.NetID[:], buf[3:6])
	copy(s.DevAddr[:], buf[6:10])
	s.DLSettings = buf[10]
	s.RXDelay = buf[11]

	if len(buf) > 16 {
		copy(s.CFList[:], buf[12:28])
	}
	rxMic := buf[len(buf)-4:]

	dataMic := []byte{}
	dataMic = append(dataMic, phyPload[0])
	dataMic = append(dataMic, o.appNonce[:]...)
	dataMic = append(dataMic, o.NetID[:]...)
	dataMic = append(dataMic, s.DevAddr[:]...)
	dataMic = append(dataMic, s.DLSettings)
	dataMic = append(dataMic, s.RXDelay)
	dataMic = append(dataMic, s.CFList[:]...)
	computedMic := genPayloadMIC(dataMic[:], o.AppKey)
	if !bytes.Equal(computedMic[:], rxMic[:]) {
		return ErrInvalidMic
	}
	// Generate NwkSKey
	// NwkSKey = aes128_encrypt(AppKey, 0x01|AppNonce|NetID|DevNonce|pad16)
	sKey := []byte{}
	sKey = append(sKey, 0x01)
	sKey = append(sKey, o.appNonce[:]...)
	sKey = append(sKey, o.NetID[:]...)
	sKey = append(sKey, o.devNonce[:]...)
	for i := 0; i < 7; i++ {
		sKey = append(sKey, 0x00) // PAD to 16
	}
	block.Encrypt(buf, sKey)
	copy(s.NwkSKey[:], buf[0:16])

	// Generate AppSKey
	// AppSKey = aes128_encrypt(AppKey, 0x02|AppNonce|NetID|DevNonce|pad16)
	sKey[0] = 0x02
	block.Encrypt(buf, sKey)
	copy(s.AppSKey[:], buf[0:16])

	// Reset counters
	s.FCntDown = 0
	s.FCntUp = 0

	return nil
}
