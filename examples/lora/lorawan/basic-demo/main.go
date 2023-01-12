// Simple code for connecting to Lorawan network and uploading sample payload
package main

import (
	"errors"
	"strconv"
	"time"

	"tinygo.org/x/drivers/examples/lora/lorawan/common"
	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/lora/lorawan"
	"tinygo.org/x/drivers/lora/lorawan/region"
)

const (
	debug                       = true
	LORAWAN_JOIN_TIMEOUT_SEC    = 180
	LORAWAN_RECONNECT_DELAY_SEC = 15
	LORAWAN_UPLINK_DELAY_SEC    = 60
)

var (
	radio   lora.Radio
	session *lorawan.Session
	otaa    *lorawan.Otaa
)

func loraConnect() error {
	start := time.Now()
	var err error
	for time.Since(start) < LORAWAN_JOIN_TIMEOUT_SEC*time.Second {
		println("Trying to join network")
		err = lorawan.Join(otaa, session)
		if err == nil {
			println("Connected to network !")
			return nil
		}
		println("Join error:", err, "retrying in", LORAWAN_RECONNECT_DELAY_SEC, "sec")
		time.Sleep(time.Second * LORAWAN_RECONNECT_DELAY_SEC)
	}

	err = errors.New("Unable to join Lorawan network")
	println(err.Error())
	return err
}

func failMessage(err error) {
	println("FATAL:", err)
	for {
	}
}

func main() {
	println("*** Lorawan basic join and uplink demo ***")

	// Board specific Lorawan initialization
	var err error
	radio, err = common.SetupLora()
	if err != nil {
		failMessage(err)
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

	lorawan.UseRegionSettings(region.EU868())

	// Configure AppEUI, DevEUI, APPKey
	setLorawanKeys()

	if debug {
		println("main: Network joined")
		println("main: DevEui, " + otaa.GetDevEUI())
		println("main: AppEui, " + otaa.GetAppEUI())
		println("main: DevAddr, " + session.GetDevAddr())
	}

	// Try to connect Lorawan network
	if err := loraConnect(); err != nil {
		failMessage(err)
	}

	if debug {
		println("main: NetID, " + otaa.GetNetID())
		println("main: NwkSKey, " + session.GetNwkSKey())
		println("main: AppSKey, " + session.GetAppSKey())
		println("main: Done")
	}
	// Try to periodicaly send an uplink sample message
	upCount := 1
	for {
		payload := "Hello TinyGo #" + strconv.Itoa(upCount)

		if err := lorawan.SendUplink([]byte(payload), session); err != nil {
			println("Uplink error:", err)
		} else {
			println("Uplink success, msg=", payload)
		}

		println("Sleeping for", LORAWAN_UPLINK_DELAY_SEC, "sec")
		time.Sleep(time.Second * LORAWAN_UPLINK_DELAY_SEC)
		upCount++
	}
}
