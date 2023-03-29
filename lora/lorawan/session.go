package lorawan

import (
	"crypto/aes"
	"encoding/binary"
	"encoding/hex"
	"math"
)

// Session is used to store session data of a LoRaWAN session
type Session struct {
	NwkSKey    [16]uint8
	AppSKey    [16]uint8
	DevAddr    [4]uint8
	FCntDown   uint32
	FCntUp     uint32
	CFList     [16]uint8
	RXDelay    uint8
	DLSettings uint8
}

// SetDevAddr configures the Session DevAddr
func (s *Session) SetDevAddr(devAddr []uint8) error {
	if len(devAddr) != 4 {
		return ErrInvalidDevAddrLength
	}

	copy(s.DevAddr[:], devAddr)

	return nil
}

// GetDevAddr returns the Session DevAddr
func (s *Session) GetDevAddr() string {
	return hex.EncodeToString(s.DevAddr[:])
}

// SetNwkSKey configures the Session NwkSKey
func (s *Session) SetNwkSKey(nwkSKey []uint8) error {
	if len(nwkSKey) != 16 {
		return ErrInvalidNwkSKeyLength
	}

	copy(s.NwkSKey[:], nwkSKey)

	return nil
}

// GetNwkSKey returns the Session NwkSKey
func (s *Session) GetNwkSKey() string {
	return hex.EncodeToString(s.NwkSKey[:])
}

// SetAppSKey configures the Session AppSKey
func (s *Session) SetAppSKey(appSKey []uint8) error {
	if len(appSKey) != 16 {
		return ErrInvalidAppSKeyLength
	}

	copy(s.AppSKey[:], appSKey)

	return nil
}

// GetAppSKey returns the Session AppSKey
func (s *Session) GetAppSKey() string {
	return hex.EncodeToString(s.AppSKey[:])
}

// GenMessage generates an uplink message.
func (s *Session) GenMessage(dir uint8, payload []uint8) ([]uint8, error) {
	var buf []uint8
	buf = append(buf, 0b01000000) // FHDR Unconfirmed up
	buf = append(buf, s.DevAddr[:]...)

	// FCtl : No ADR, No RFU, No ACK, No FPending, No FOpt
	buf = append(buf, 0x00)

	// FCnt Up
	buf = append(buf, uint8(s.FCntUp&0xFF), uint8((s.FCntUp>>8)&0xFF))

	// FPort=1
	buf = append(buf, 0x01)

	fCnt := uint32(0)
	if dir == 0 {
		fCnt = s.FCntUp
		s.FCntUp++
	} else {
		fCnt = s.FCntDown
	}
	data, err := s.genFRMPayload(dir, fCnt, payload, false)
	if err != nil {
		return nil, err
	}
	buf = append(buf, data[:]...)

	mic := calcMessageMIC(buf, s.NwkSKey, dir, s.DevAddr[:], fCnt, uint8(len(buf)))
	buf = append(buf, mic[:]...)

	return buf, nil
}

func (s *Session) genFRMPayload(dir uint8, fCnt uint32, payload []byte, isFOpts bool) ([]byte, error) {
	k := len(payload) / aes.BlockSize
	if len(payload)%aes.BlockSize != 0 {
		k++
	}
	if k > math.MaxUint8 {
		return nil, ErrFrmPayloadTooLarge
	}
	encrypted := make([]byte, 0, k*16)
	cipher, err := aes.NewCipher(s.AppSKey[:])
	if err != nil {
		panic(err)
	}

	var a [aes.BlockSize]byte
	a[0] = 0x01
	a[5] = dir
	copy(a[6:10], s.DevAddr[:])
	binary.LittleEndian.PutUint32(a[10:14], fCnt)
	var ss [aes.BlockSize]byte
	var b [aes.BlockSize]byte
	for i := uint8(0); i < uint8(k); i++ {
		copy(b[:], payload[i*aes.BlockSize:])
		if !isFOpts {
			a[15] = i + 1
		}
		cipher.Encrypt(ss[:], a[:])
		for j := 0; j < aes.BlockSize; j++ {
			b[j] = b[j] ^ ss[j]
		}
		encrypted = append(encrypted, b[:]...)
	}
	return encrypted[:len(payload)], nil
}
