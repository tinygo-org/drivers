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

var debug string

const (
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

	// Connect the lorawan with the Lora Radio device.
	lorawan.UseRadio(radio)

	lorawan.UseRegionSettings(region.EU868())

	// Configure AppEUI, DevEUI, APPKey, and public/private Lorawan Network
	setLorawanKeys()

	if debug != "" {
		println("main: Network joined")
		println("main: DevEui, " + otaa.GetDevEUI())
		println("main: AppEui, " + otaa.GetAppEUI())
		println("main: DevAddr, " + otaa.GetAppKey())
	}

	// Try to connect Lorawan network
	if err := loraConnect(); err != nil {
		failMessage(err)
	}

	if debug != "" {
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
