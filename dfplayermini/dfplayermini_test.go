package DFPlayerMini

import (
	"testing"
	"time"

	"go.bug.st/serial"
)

const (
	minTrackPlaybackTime   = time.Duration(time.Second * 3)
	trackPlaytimeIncrement = time.Duration(time.Second)
	serialPortPath         = "/dev/ttyUSB1"
)

func TestPlayerReset(t *testing.T) {
	sp, err := serial.Open(serialPortPath, &serial.Mode{BaudRate: 9600, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit})
	if err != nil {
		t.Fail()
		return
	}

	sp.SetReadTimeout(time.Millisecond * 10)
	defer sp.Close()

	player := New(sp, DebugLevel1)

	time.Sleep(time.Millisecond * 500)

	player.Reset()

	player.SelectPlaybackSource(PlaybackSourceSD)

	player.WaitStorageReady()
}

func TestGetSDTrackCount(t *testing.T) {
	sp, err := serial.Open(serialPortPath, &serial.Mode{BaudRate: 9600, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit})
	if err != nil {
		t.Fail()
		return
	}

	sp.SetReadTimeout(time.Millisecond * 10)
	defer sp.Close()

	player := New(sp, DebugQuiet)

	time.Sleep(time.Millisecond * 500)

	player.SelectPlaybackSource(PlaybackSourceSD)

	tracks, ok := player.GetSDTrackCount()
	if !ok || tracks == 0 {
		t.Fail()
	}

	player.Sleep()
}

func TestFolderEnumeration(t *testing.T) {
	sp, err := serial.Open(serialPortPath, &serial.Mode{BaudRate: 9600, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit})
	if err != nil {
		t.Fail()
		return
	}

	sp.SetReadTimeout(time.Millisecond * 10)
	defer sp.Close()

	player := New(sp, DebugQuiet)
	player.SetMaxTestTrackRuntime(time.Second * 20)

	time.Sleep(time.Millisecond * 500)

	player.SelectPlaybackSource(PlaybackSourceSD)

	folders, totalTracks := player.BuildFolderPlaylist()

	if len(folders) == 0 || totalTracks == 0 {
		t.Fail()
	}

	player.Sleep()
}

func TestPlayNextTrack(t *testing.T) {
	sp, err := serial.Open(serialPortPath, &serial.Mode{BaudRate: 9600, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit})
	if err != nil {
		t.Fail()
		return
	}

	sp.SetReadTimeout(time.Millisecond * 10)
	defer sp.Close()

	player := New(sp, DebugLevel1)
	player.SetMaxTestTrackRuntime(time.Second * 20)

	time.Sleep(time.Millisecond * 500)

	player.SelectPlaybackSource(PlaybackSourceSD)
	player.SetVolume(3)
	player.SetEQ(EqBass)
	player.StopRepeatPlayback()

	player.StartDAC()
	player.SetAmplificationGain(true, 2)

	player.PlayNextTrack()

	for {
		status := player.CheckTrackStatus(time.Second, time.Second*3)
		switch status {
		case SdTrackFinished:
			goto EXIT
		case ErrorTrackNotFound:
			panic("Track not found")
		case MediaOut:
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

func TestPlayFolderTrack(t *testing.T) {
	sp, err := serial.Open(serialPortPath, &serial.Mode{BaudRate: 9600, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit})
	if err != nil {
		t.Fail()
		return
	}

	sp.SetReadTimeout(time.Millisecond * 10)
	defer sp.Close()

	player := New(sp, DebugLevel1)
	player.SetMaxTestTrackRuntime(time.Second * 20)

	time.Sleep(time.Millisecond * 500)

	player.SelectPlaybackSource(PlaybackSourceSD)
	player.SetVolume(3)
	player.SetEQ(EqBass)
	player.StopRepeatPlayback()

	player.StartDAC()
	player.SetAmplificationGain(true, 2)

	player.PlayFolderTrack(1, 1)

	for {
		status := player.CheckTrackStatus(time.Second, time.Second*3)
		switch status {
		case SdTrackFinished:
			goto EXIT
		case ErrorTrackNotFound:
			panic("Track not found")
		case MediaOut:
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

func TestPlayAdvertFolderTrack(t *testing.T) {
	sp, err := serial.Open(serialPortPath, &serial.Mode{BaudRate: 9600, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit})
	if err != nil {
		t.Fail()
		return
	}

	sp.SetReadTimeout(time.Millisecond * 10)
	defer sp.Close()

	player := New(sp, DebugLevel1)
	player.SetMaxTestTrackRuntime(time.Second * 20)

	time.Sleep(time.Millisecond * 500)

	player.SelectPlaybackSource(PlaybackSourceSD)
	player.SetVolume(3)
	player.SetEQ(EqBass)
	player.StopRepeatPlayback()

	player.StartDAC()
	player.SetAmplificationGain(true, 2)

	player.PlayMP3FolderTrack(1)
	time.Sleep(time.Second * 3)
	player.PlayAdvertFolder(1)

	for {
		status := player.CheckTrackStatus(time.Second, time.Second*3)
		switch status {
		case SdTrackFinished:
			goto EXIT
		case ErrorTrackNotFound:
			panic("Track not found")
		case MediaOut:
			panic("Media out")
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}

EXIT:
	player.StopAdvert()
	player.Stop()
	player.StopDAC()
	player.Sleep()
}

func TestPlay3KFolderTrack(t *testing.T) {
	sp, err := serial.Open(serialPortPath, &serial.Mode{BaudRate: 9600, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit})
	if err != nil {
		t.Fail()
		return
	}

	sp.SetReadTimeout(time.Millisecond * 10)
	defer sp.Close()

	player := New(sp, DebugLevel1)
	player.SetMaxTestTrackRuntime(time.Second * 20)

	time.Sleep(time.Millisecond * 500)

	player.SelectPlaybackSource(PlaybackSourceSD)
	player.SetVolume(3)
	player.SetEQ(EqBass)
	player.StopRepeatPlayback()

	player.StartDAC()
	player.SetAmplificationGain(true, 2)

	player.Play3KFolderTrack(10, 3000)

	for {
		status := player.CheckTrackStatus(time.Second, time.Second*3)
		switch status {
		case SdTrackFinished:
			goto EXIT
		case ErrorTrackNotFound:
			panic("Track not found")
		case MediaOut:
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
