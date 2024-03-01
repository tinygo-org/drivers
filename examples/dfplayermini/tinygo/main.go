package main

import (
	"fmt"
	"machine"
	"time"

	df "tinygo.org/x/drivers/dfplayermini"
)

const (
	mp3Tx = machine.GPIO4
	mp3Rx = machine.GPIO5
)

func main() {
	mp3Serial := machine.UART1
	err := mp3Serial.Configure(machine.UARTConfig{BaudRate: 9600, TX: mp3Tx, RX: mp3Rx})
	if err != nil {
		println(fmt.Sprintf("UART config failed. Err: %v", err), 0)
		panic(err)
	}

	time.Sleep(time.Millisecond * 500)

	player := df.New(mp3Serial, df.DebugQuiet)

	player.Reset()
	player.SelectPlaybackSource(df.PlaybackSourceSD)
	player.SetVolume(3)
	player.SetEQ(df.EqBass)
	player.StopRepeatPlayback()

	player.StartDAC()
	player.SetAmplificationGain(true, 2)

	player.PlayFolderTrack(1, 1)

	for {
		status := player.CheckTrackStatus(time.Second, time.Second*3)
		switch status {
		case df.SdTrackFinished:
			goto EXIT
		case df.ErrorTrackNotFound:
			panic("Track not found")
		case df.MediaOut:
			panic("Media out")
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}

EXIT:
	player.Stop()
	player.StopDAC()
	player.Sleep()
}
