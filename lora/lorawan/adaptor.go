package lorawan

import (
	"errors"

	"tinygo.org/x/drivers/lora"
)

const LORA_RXTX_TIMEOUT = 1000

var (
	ActiveRadio lora.Radio
)

func UseRadio(r lora.Radio) {
	if ActiveRadio != nil {
		panic("lorawan.ActiveRadio is already set")
	}
	ActiveRadio = r
}

func Join(otaa *Otaa, session *Session) error {
	var resp []uint8

	if ActiveRadio == nil {
		return errors.New("no LoRa radio attached")
	}

	otaa.Init()

	// Send join packet
	payload, err := otaa.GenerateJoinRequest()
	if err != nil {
		return err
	}

	ActiveRadio.SetCrc(true)
	ActiveRadio.SetIqMode(0) // IQ Standard
	ActiveRadio.Tx(payload, LORA_RXTX_TIMEOUT)
	if err != nil {
		return err
	}

	// Wait for JoinAccept
	//println("lorawan: Wait for JOINACCEPT for 10s")
	ActiveRadio.SetIqMode(1) // IQ Inverted
	for i := 0; i < 15; i++ {
		resp, err = ActiveRadio.Rx(LORA_RXTX_TIMEOUT)
		if err != nil {
			return err
		}
		if resp != nil {
			break
		}
	}
	if resp == nil {
		return errors.New("no JoinAccept packet received")
	}
	//println("lorawan: Received packet: len=", len(resp), "payload=", bytesToHexString(resp))

	err = otaa.DecodeJoinAccept(resp, session)
	if err != nil {
		return err
	}
	// println("lorawan: Valid JOINACCEPT, now connected")
	// println("lorawan: |  DevAddr: ", bytesToHexString(r.Session.DevAddr[:]), " (LSB)")
	// println("lorawan: |  NetID  : ", bytesToHexString(r.Otaa.NetID[:]))
	// println("lorawan: |  NwkSKey: ", bytesToHexString(r.Session.NwkSKey[:]))
	// println("lorawan: |  AppSKey: ", bytesToHexString(r.Session.AppSKey[:]))

	return nil
}

func SendUplink() error {
	return nil
}

func ListenDownlink() error {
	return nil
}
