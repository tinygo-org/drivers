// Simple code for connecting to Lorawan network and uploading sample payload

package main

import (
	"time"

	"tinygo.org/x/drivers/examples/lora/lorawan/common"
	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/lora/lorawan"
)

const (
	LORAWAN_RECONNECT_DELAY_SEC = 60
)

var (
	radio   lora.Radio
	session *lorawan.Session
	otaa    *lorawan.Otaa

	lorawanConnected bool
)

// loraConnect() will loop until we're connected to Lorawan network
func loraConnect() {
	for {
		for !lorawanConnected {
			println("Starting Lorawan Join sequence")
			err := lorawan.Join(otaa, session)
			if err != nil {
				println("loraConnect: Join error:", err)
				println("loraConnect: Wait 300 sec")
				time.Sleep(time.Second * LORAWAN_RECONNECT_DELAY_SEC)
			} else {
				println("loraConnect: Connected !")
				lorawanConnected = true
			}
		}
		time.Sleep(time.Second * LORAWAN_RECONNECT_DELAY_SEC)
	}
}

func main() {
	println("Lorawan Simple Demo")

	// Board specific Lorawan initialization
	var err error
	radio, err = common.SetupLora()
	if err != nil {
		println("FATAL:", err.Error())
		for {
		}
	}

	// Required for LoraWan operations
	session = &lorawan.Session{}
	otaa = &lorawan.Otaa{}

	// Initial Lora modulation configuration
	loraConf := lora.Config{
		Freq:           868100000,
		Bw:             lora.Bandwidth_125_0,
		Sf:             lora.SpreadingFactor9,
		Cr:             lora.CodingRate4_7,
		HeaderType:     lora.HeaderExplicit,
		Preamble:       12,
		Ldr:            lora.LowDataRateOptimizeOff,
		Iq:             lora.IQStandard,
		Crc:            lora.CRCOn,
		SyncWord:       lora.SyncPublic,
		LoraTxPowerDBm: 20,
	}
	radio.LoraConfig(loraConf)

	// Connect the lorawan with the Lora Radio device.
	lorawan.UseRadio(radio)

	// Configure AppEUI, DevEUI, APPKey
	setLorawanKeys()

	// Background lorawan connection handler
	go loraConnect()

	payload := []byte("Hello Tinygo")

	for {
		if lorawanConnected {
			println("Sending uplink message")
			if err := lorawan.SendUplink(payload, session); err != nil {
				println("Uplink error:", err)
			}
			// Do something like uploading datas.
		}

		time.Sleep(time.Second * 30)
	}
}
