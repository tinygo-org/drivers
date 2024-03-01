package main

import (
	"fmt"
	"time"

	"go.bug.st/serial"
	df "tinygo.org/x/drivers/dfplayermini"
)

const (
	minTrackPlaybackTime   = time.Duration(time.Second * 3)
	trackPlaytimeIncrement = time.Duration(time.Second)
	serialPortPath         = "/dev/ttyUSB1"
)

func main() {
	sp, err := serial.Open(serialPortPath, &serial.Mode{BaudRate: 9600, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit})
	if err != nil {
		println(fmt.Sprintf("serial.Open() failed: %v", err))
		panic(err)
	}

	sp.SetReadTimeout(time.Millisecond * 10)
	defer sp.Close()

	player := df.New(sp, df.DebugQuiet)

	time.Sleep(time.Millisecond * 500)

	for {
	INIT:
		player.SelectPlaybackSource(df.PlaybackSourceSD)
		player.Discard()
		_, ok := player.GetSDTrackCount()
		if !ok {
			time.Sleep(time.Second)
			goto INIT
		}

		player.SetVolume(0)
		player.StopDAC()

		time.Sleep(time.Millisecond * 500)

		folders, totalTracks := player.BuildFolderPlaylist()
		println(fmt.Sprintf("SD card numerical folders: %02d, containing %02d tracks", len(folders), totalTracks))

		player.SetDebugLevel(df.DebugLevel1)

		player.StartDAC()
		player.SetAmplificationGain(true, 2)
		player.SetVolume(1)
		player.SetEQ(df.EqRock)
		player.StopRepeatPlayback()

		for folder, folderTracks := range folders {
			currentTrack := 0
			for {
				currentTrack++
				if currentTrack > int(folderTracks) {
					break
				}

				player.PlayFolderTrack(folder, uint8(currentTrack))

				for {
					status := player.CheckTrackStatus(time.Second, time.Second*3)
					switch status {
					case df.SdTrackFinished:
						goto NEXT

					case df.ErrorTrackNotFound:
						goto NEXT

					case df.MediaOut:
						goto INIT

					case df.TrackPlaying:

					default:
						if player.GetDebugLevel() > 2 {
							println(fmt.Sprintf("Unexpected status: 0x%0X", status))
						}
						time.Sleep(time.Millisecond * 100)
					}
				}
			NEXT:
			}
		}
		break
	}

	player.SetVolume(0)
	player.Stop()
	player.StopDAC()
	player.Sleep()
}
