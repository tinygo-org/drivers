package DFPlayerMini

// Package DFPlayerMini provides a driver for the MP3 player by DFRobot
//
// Author: Fabien Royer
//
// Datasheet:
// https://github.com/DFRobot/DFRobotDFPlayerMini/blob/master/doc/FN-M16P%2BEmbedded%2BMP3%2BAudio%2BModule%2BDatasheet.pdf

import (
	"fmt"
	"strings"
	"time"
)

const (
	FrameStartByteOffset = iota
	FrameVersionOffset
	FrameLengthOffset
	FrameCommandOffset
	FrameFeedbackOffset
	FrameParamMSBOffset
	FrameParamLSBOffset
	FrameChecksumMSBOffset
	FrameChecksumLSBOffset
	FrameEndByteOffset
	FrameSize = FrameEndByteOffset + 1

	StartByte           = 0x7E
	Version             = 0xFF
	Length              = 0x06
	FeedbackRequired    = 0x01
	FeedbackNotRequired = 0x00
	EndByte             = 0xEF

	EqNormal  = 0x00
	EqPop     = 0x01
	EqRock    = 0x02
	EqJazz    = 0x03
	EqClassic = 0x04
	EqBass    = 0x05

	PlayNextTrack     = 0x01
	PlayPreviousTrack = 0x02
	PlayRootTrack     = 0x03
	VolumeUp          = 0x04
	VolumeDown        = 0x05
	Volume            = 0x06
	Eq                = 0x07
	PlaybackMode      = 0x08
	PlaybackSource    = 0x09
	SleepMode         = 0x0A

	Reset             = 0x0C
	ResumePlayback    = 0x0D
	Pause             = 0x0E
	PlayFolderTrack   = 0x0F
	AmplificationGain = 0x10
	RepeatPlay        = 0x11
	PlayMp3Folder     = 0x12
	PlayAdvert        = 0x13
	PlayTrack3K       = 0x14
	StopAdvert        = 0x15
	Stop              = 0x16
	RepeatFolder      = 0x17
	PlayRandomAll     = 0x18
	RepeatCurrent     = 0x19
	DAC               = 0x1A

	MediaIn  = 0x3A
	MediaOut = 0x3B

	ModuleAsleep = 0x1000

	TrackStopped = 0x00
	TrackPlaying = 0x01
	TrackPaused  = 0x02

	UsbTrackFinished = 0x3C
	SdTrackFinished  = 0x3D

	QueryStorage = 0x3F

	ErrorCondition       = 0x40
	ErrorTrackOutOfScope = 0x05
	ErrorTrackNotFound   = 0x06

	Feedback = 0x41

	GetStatus  = 0x42
	GetVolume  = 0x43
	GetEq      = 0x44
	GetVersion = 0x46

	GetUsbRootTrackCount = 0x47
	GetSdRootTrackCount  = 0x48

	GetCurrentUsbTrackCount = 0x4B
	GetCurrentSdTrack       = 0x4C

	GetFolderTrakCount = 0x4E
	GetFolderCount     = 0x4F

	PlaybackSourceUSB = 0x01
	PlaybackSourceSD  = 0x02

	delayOnWrite = time.Millisecond * 80 // 10 bytes @ 9600 kbps -> ~77.33ms transfer time

	DebugQuiet  = 0
	DebugLevel1 = 1
	DebugLevel2 = 2
)

// SerialPort abstracts the actual interface to a generic serial port
type SerialPort interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

// Wraps the DF Player Mini functions
type Device struct {
	rx                  []byte
	tx                  []byte
	tempRx              []byte
	port                SerialPort
	debug               uint
	statusErrorCount    int64
	minDelayOnWrite     time.Duration
	trackRuntime        time.Duration
	maxTestTrackRuntime time.Duration
}

// Creates and initializes a new DF Player Mini
func New(port SerialPort, dbg uint) Device {
	dev := Device{}
	dev.Init(port, dbg)
	return dev
}

// Initializes internal data structures
func (d *Device) Init(port SerialPort, dbg uint) {
	d.minDelayOnWrite = delayOnWrite
	d.debug = dbg
	d.port = port
	d.rx = make([]byte, FrameSize)
	d.tempRx = make([]byte, 1)
	d.tx = make([]byte, FrameSize)
	d.tx[FrameStartByteOffset] = StartByte
	d.tx[FrameVersionOffset] = Version
	d.tx[FrameLengthOffset] = Length
	d.tx[FrameFeedbackOffset] = FeedbackNotRequired
	d.tx[FrameEndByteOffset] = EndByte
}

func (d *Device) GetDebugLevel() uint {
	return d.debug
}

func (d *Device) SetDebugLevel(level uint) {
	d.debug = level
}

// Play the next track in the global track list maintained internally by the MP3 player
func (d *Device) PlayNextTrack() {
	d.tx[FrameCommandOffset] = PlayNextTrack
	d.setUInt16(FrameParamMSBOffset, 1)
	d.write()
}

// Play the previous track in the global track list maintained internally by the MP3 player
func (d *Device) PlayPreviousTrack() {
	d.tx[FrameCommandOffset] = PlayPreviousTrack
	d.setUInt16(FrameParamMSBOffset, 1)
	d.write()
}

// Play a track (1-3000) located at the root of the storage source
func (d *Device) PlayRootTrack(track uint16) {
	d.tx[FrameCommandOffset] = PlayRootTrack
	d.setUInt16(FrameParamMSBOffset, track)
	d.write()
}

// Play a track (1-255) located within a folder (01-99)
func (d *Device) PlayFolderTrack(folder, track uint8) {
	d.tx[FrameCommandOffset] = PlayFolderTrack
	d.tx[FrameParamMSBOffset] = folder
	d.tx[FrameParamLSBOffset] = track
	d.write()
}

// Play a track (1-3000) located within a folder (01-99)
func (d *Device) Play3KFolderTrack(folder uint8, track uint16) bool {
	if folder <= 15 {
		param := (uint16(folder) << 12) | (track & 0xfff)
		d.tx[FrameCommandOffset] = PlayTrack3K
		d.setUInt16(FrameParamMSBOffset, param)
		d.write()
		return true
	}
	return false
}

// Play a track (1-3000) located within the 'mp3' folder
func (d *Device) PlayMP3FolderTrack(track uint16) {
	d.tx[FrameCommandOffset] = PlayMp3Folder
	d.setUInt16(FrameParamMSBOffset, track)
	d.write()
}

// Play a track (1-3000) located within the 'advert' folder
func (d *Device) PlayAdvertFolder(track uint16) {
	d.tx[FrameCommandOffset] = PlayAdvert
	d.setUInt16(FrameParamMSBOffset, track)
	d.write()
}

// Stop the currently playing advert track
func (d *Device) StopAdvert() {
	d.tx[FrameCommandOffset] = StopAdvert
	d.setUInt16(FrameParamMSBOffset, 0)
	d.write()
}

// Stop all playback
func (d *Device) Stop() {
	d.tx[FrameCommandOffset] = Stop
	d.setUInt16(FrameParamMSBOffset, 0)
	d.write()
}

// Set the volume (0-31)
func (d *Device) SetVolume(vol uint8) {
	d.tx[FrameCommandOffset] = Volume
	d.setUInt16(FrameParamMSBOffset, uint16(vol))
	d.write()
}

// Set the amplification gain when using DAC
func (d *Device) SetAmplificationGain(enable bool, gain uint8) bool {
	if gain <= 31 {
		d.tx[FrameCommandOffset] = AmplificationGain
		if enable {
			d.tx[FrameParamMSBOffset] = 0x01
		} else {
			d.tx[FrameParamMSBOffset] = 0x00
		}
		d.tx[FrameParamLSBOffset] = gain
		d.write()
		return true
	}
	return false
}

// Increment the volume
func (d *Device) VolumeUp() {
	d.tx[FrameCommandOffset] = VolumeUp
	d.setUInt16(FrameParamMSBOffset, 1)
	d.write()
}

// Decrement the volume
func (d *Device) VolumeDown() {
	d.tx[FrameCommandOffset] = VolumeDown
	d.setUInt16(FrameParamMSBOffset, 1)
	d.write()
}

// Set the equalizer to a specific envelope (0-5)
func (d *Device) SetEQ(eq uint8) {
	if eq <= 5 {
		d.tx[FrameCommandOffset] = Eq
		d.setUInt16(FrameParamMSBOffset, uint16(eq))
		d.write()
	}
}

// Play a track in a loop
func (d *Device) LoopTrack(track uint16) {
	d.tx[FrameCommandOffset] = PlaybackMode
	d.setUInt16(FrameParamMSBOffset, track)
	d.write()
}

// Specify which data source should be used by the player (USB storage or SD card)
func (d *Device) SelectPlaybackSource(src uint8) {
	d.tx[FrameCommandOffset] = PlaybackSource
	d.setUInt16(FrameParamMSBOffset, uint16(src))
	d.write()
	time.Sleep(time.Millisecond * 200)
}

// Enter sleep mode
func (d *Device) Sleep() {
	d.tx[FrameCommandOffset] = SleepMode
	d.setUInt16(FrameParamMSBOffset, 1)
	d.write()
}

// Player reset
func (d *Device) Reset() {
	d.tx[FrameCommandOffset] = Reset
	d.setUInt16(FrameParamMSBOffset, 0)
	d.write()
}

// Resume playback
func (d *Device) Resume() {
	d.tx[FrameCommandOffset] = ResumePlayback
	d.setUInt16(FrameParamMSBOffset, 1)
	d.write()
}

// Pause playback
func (d *Device) Pause() {
	d.tx[FrameCommandOffset] = Pause
	d.setUInt16(FrameParamMSBOffset, 1)
	d.write()
}

// Repeat playback indefinitely
func (d *Device) StartRepeatPlayback() {
	d.tx[FrameCommandOffset] = RepeatPlay
	d.setUInt16(FrameParamMSBOffset, 1)
	d.write()
}

// Stop repeating playback
func (d *Device) StopRepeatPlayback() {
	d.tx[FrameCommandOffset] = RepeatPlay
	d.setUInt16(FrameParamMSBOffset, 0)
	d.write()
}

// Repeat playback of a specific folder
func (d *Device) RepeatFolder(folder uint16) {
	d.tx[FrameCommandOffset] = RepeatFolder
	d.setUInt16(FrameParamMSBOffset, folder)
	d.write()
}

// Random playback across all tracks available
func (d *Device) RandomPlaybackAll() {
	d.tx[FrameCommandOffset] = PlayRandomAll
	d.setUInt16(FrameParamMSBOffset, 0)
	d.write()
}

// Repeat the current track
func (d *Device) StartRepeatCurrentTrack() {
	d.tx[FrameCommandOffset] = RepeatCurrent
	d.setUInt16(FrameParamMSBOffset, 0)
	d.write()
}

// Stop repeating the current track
func (d *Device) StopRepeatCurrentTrack() {
	d.tx[FrameCommandOffset] = RepeatCurrent
	d.setUInt16(FrameParamMSBOffset, 1)
	d.write()
}

// Enable the DAC output to an amplifier
func (d *Device) StartDAC() {
	d.tx[FrameCommandOffset] = DAC
	d.setUInt16(FrameParamMSBOffset, 0)
	d.write()
}

// Stop the DAC output to an amplifier
func (d *Device) StopDAC() {
	d.tx[FrameCommandOffset] = DAC
	d.setUInt16(FrameParamMSBOffset, 1)
	d.write()
}

// Query the status of the player (playing, stopped, media removed, etc.)
// uint8: status type identifier (a.k.a command)
// uint16: MSB and LSB parameter bytes
// bool: success or failure querying status
func (d *Device) QueryStatus() (uint8, uint16, bool) {
	cmd, result, ok := d.query(GetStatus, 0, 0)
	if ok {
		return cmd, result, true
	}
	return 0, 0, false
}

// Get the online storage type
func (d *Device) GetOnlineStorage() (uint16, bool) {
	_, result, ok := d.query(QueryStorage, 0, 0)
	if ok {
		return result, true
	}
	return 0, false
}

// Get the volume value
func (d *Device) GetVolume() (uint8, bool) {
	_, result, ok := d.query(GetVolume, 0, 0)
	if ok {
		return uint8(result), true
	}
	return 0, false
}

// Get the equalizer setting
func (d *Device) GetEQ() (uint8, bool) {
	_, result, ok := d.query(GetEq, 0, 0)
	if ok {
		return uint8(result), true
	}
	return 0, false
}

// Get the version of the MP3 player module
func (d *Device) GetVersion() (uint16, bool) {
	_, result, ok := d.query(GetVersion, 0, 0)
	if ok {
		return result, true
	}
	return 0, false
}

// Get the total track count in the USB storage
func (d *Device) GetUSBTrackCount() (uint16, bool) {
	_, result, ok := d.query(GetUsbRootTrackCount, 0, 0)
	if ok {
		return result, true
	}
	return 0, false
}

// Get the total track count in the SD card storage
func (d *Device) GetSDTrackCount() (uint16, bool) {
	_, result, ok := d.query(GetSdRootTrackCount, 0, 0)
	if ok {
		return result, true
	}
	return 0, false
}

// Get the current internal track number being played from the USB storage
// Note: the track numbers being returned are internal to the MP3 player's track list.
func (d *Device) GetCurrentUSBtrack() (uint16, bool) {
	_, result, ok := d.query(GetCurrentUsbTrackCount, 0, 0)
	if ok {
		return result, true
	}
	return 0, false
}

// Get the current internal track number being played from the SD card storage
// Note: the track numbers  being returned are internal to the MP3 player's track list.
func (d *Device) GetCurrentSDtrack() (uint16, bool) {
	_, result, ok := d.query(GetCurrentSdTrack, 0, 0)
	if ok {
		return result, true
	}
	return 0, false
}

// Get the track count within a folder (1-99)
func (d *Device) GetFolderTrackCount(folder uint8) (uint16, bool) {
	cmd, result, ok := d.query(GetFolderTrakCount, 0, folder)
	if ok && cmd != ErrorCondition {
		return result, true
	}
	return 0, false
}

// Get the folder count at the root of the storage device.
// The total folder count include all numeric folders (01-99) and the 'mp3' and 'advert' folders
func (d *Device) GetFolderCount() (uint16, bool) {
	_, result, ok := d.query(GetFolderCount, 0, 0)
	if ok {
		return result, true
	}
	return 0, false
}

// Discard any pending bytes waiting to be read
func (d *Device) Discard() {
	for {
		n, err := d.port.Read(d.tempRx)
		if n == 0 || err != nil {
			break
		}
	}
}

// Enumerate the root storage folders (01-99) and retrieve the track count within each one
// Returns a map of folder -> track count
func (d *Device) BuildFolderPlaylist() (map[uint8]uint16, uint16) {
	d.Discard()
	var total uint16
	pl := make(map[uint8]uint16, 0)
	var folder uint8
	for folder = 1; folder < 100; folder++ {
		files, ok := d.GetFolderTrackCount(uint8(folder))
		if ok {
			total = total + files
			pl[folder] = files
		}
	}
	d.Discard()
	return pl, total
}

// Check on the current playback status
// Returns ErrorCondition in case a status query fails.
// Returns ErrorTrackNotFound when a track or folder is not found.
// Returns MediaOut when the storage is ejected.
// Returns SdTrackFinished when a track complete playback
// Return TrackPlaying when a track is in progress
func (d *Device) CheckTrackStatus(trackPlaytimeIncrement, minTrackPlaybackTime time.Duration) uint {
	cmd, param, ok := d.QueryStatus()
	if ok {
		switch cmd {
		case ErrorCondition:
			if param == ErrorTrackOutOfScope || param == ErrorTrackNotFound {
				return ErrorTrackNotFound
			}

		case MediaOut:
			return MediaOut

		case SdTrackFinished:
			if d.trackRuntime > minTrackPlaybackTime {
				if d.debug > 0 {
					println(fmt.Sprintf("SD card track #%04d finished playing", param))
				}
				d.trackRuntime = time.Duration(time.Second * 0)
				return SdTrackFinished
			}

		case GetStatus:
			if (param & 0x00FF) == TrackPlaying {
				d.trackRuntime = d.trackRuntime + trackPlaytimeIncrement
				if d.maxTestTrackRuntime > 0 && d.trackRuntime > d.maxTestTrackRuntime {
					return SdTrackFinished
				}
				time.Sleep(trackPlaytimeIncrement)
				return TrackPlaying
			}
		}
	} else {
		d.statusErrorCount++
		if d.debug > 0 {
			println(fmt.Sprintf("QueryStatus() error count: %02d", d.statusErrorCount))
		}
	}
	return ErrorCondition
}

// Set for a maximum track playback duration during tests
func (d *Device) SetMaxTestTrackRuntime(max time.Duration) {
	d.maxTestTrackRuntime = max
	d.trackRuntime = 0
}

// Wait for the device to complete its initialization after a reset
func (d *Device) WaitStorageReady() {
	d.Discard()
	for {
		_, ok := d.GetOnlineStorage()
		if ok {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}

// General function used to query the MP3 playwer
func (d *Device) query(cmd, msb, lsb uint8) (uint8, uint16, bool) {
	d.tx[FrameCommandOffset] = cmd
	d.tx[FrameParamMSBOffset] = msb
	d.tx[FrameParamLSBOffset] = lsb
	d.computeChecksum(d.tx)
	d.write()

	respCmd, param, valid := d.readMp3Response()

	if valid {
		return respCmd, param, valid
	}
	return 0, 0, false
}

// Internal function used to write data to the MP3 player's serial interface
func (d *Device) write() {
	sum := d.computeChecksum(d.tx)
	d.setUInt16(FrameChecksumMSBOffset, sum)
	if d.debug >= 2 {
		println(fmt.Sprintf("tx: %02x", d.tx))
	}
	d.port.Write(d.tx)
	time.Sleep(d.minDelayOnWrite)
}

// Internal function used to read responses sent by the MP3 player
// This function reads a single byte at a time to remain compatible with TinyGo
func (d *Device) readMp3Response() (uint8, uint16, bool) {
	for i := range d.rx {
		d.rx[i] = 0
	}
	byteCount := 0
	for ; byteCount < FrameSize; byteCount++ {
		_, err := d.port.Read(d.tempRx)
		if err != nil {
			if d.debug > 0 {
				println(fmt.Sprintf("rx err: %s", err))
			}
			return 0, 0, false
		}
		d.rx[byteCount] = d.tempRx[0]
	}

	if !d.isRxBufferEmpty() && byteCount == FrameSize {
		if d.debug >= 2 {
			println(fmt.Sprintf("rx: %02x", d.rx))
		}
		if d.validateChecksum() {
			if d.debug > DebugQuiet {
				d.decodeResponse()
			}
			return d.rx[FrameCommandOffset], d.getUInt16(FrameParamMSBOffset), true
		} else {
			if d.debug > 0 {
				println(fmt.Sprintf("bad checksum rx: %02x", d.rx))
			}
		}
	}

	return 0, 0, false
}

// Determines if the received buffer only contains 'zero' bytes
func (d *Device) isRxBufferEmpty() bool {
	for _, v := range d.rx {
		if v != 0 {
			return false
		}
	}
	return true
}

// Computes a checksum derived from the received data and compares it to the checksum computed by the MP3 player.
// Return 'true' when both checksums match. Note that the MP3 player can return 'shifted' bytes in responses (bug?).
// This function attempts to re-organize a 'shifted' response into a proper one before validation.
func (d *Device) validateChecksum() bool {
	if d.rx[FrameStartByteOffset] != StartByte || d.rx[FrameVersionOffset] != Version || d.rx[FrameLengthOffset] != Length || d.rx[FrameEndByteOffset] != EndByte {
		var frameSb bool
		var frameVer bool
		var frameLen bool
		var frameEb bool
		var posSB int
		var v byte
		var idx int
		for idx, v = range d.rx {
			switch v {
			case StartByte:
				posSB = idx
				frameSb = true
			case Version:
				frameVer = true
			case Length:
				frameLen = true
			case EndByte:
				frameEb = true
			}
		}
		if frameSb && frameVer && frameLen && frameEb {
			if d.debug >= 2 {
				print(fmt.Sprintf("Shifting rx frame %0X -> ", d.rx))
			}
			newRx := make([]byte, FrameSize)
			newRxIdx := 0
			for i := posSB; i < FrameSize; i++ {
				newRx[newRxIdx] = d.rx[i]
				newRxIdx++
			}
			if posSB > 0 {
				for i := 0; i < posSB; i++ {
					newRx[newRxIdx] = d.rx[i]
					newRxIdx++
				}
			}
			d.rx = newRx
			if d.debug >= 2 {
				println(fmt.Sprintf("%0X", d.rx))
			}
		}
	}

	csum := d.computeChecksum(d.rx)
	rsum := d.getUInt16(FrameChecksumMSBOffset)
	if d.debug >= 2 {
		println(fmt.Sprintf("checksum calc:0x%04X rx:0x%04X", csum, rsum))
	}
	return csum == rsum
}

// Computes a checksum from the received buffer
func (d *Device) computeChecksum(a []byte) uint16 {
	var sum uint16
	for i := 1; i <= 6; i++ {
		sum = sum + uint16(a[i])
	}
	sum = 0xffff - (sum) + 1
	return sum
}

// Set a 16 bit value in the transmission buffer
func (d *Device) setUInt16(offset uint8, i uint16) {
	d.tx[offset] = uint8(i >> 8)
	d.tx[offset+1] = uint8(i & 0xff)
}

// Get a 16 bit value from the receive buffer
func (d *Device) getUInt16(offset uint8) uint16 {
	var i uint16
	i = uint16(d.rx[offset])
	i = i << 8
	i = i | uint16(d.rx[offset+1])
	return i
}

// Attempt to decode the response sent by the MP3 player.
// Returns 'false' if the response contains an error condition.
func (d *Device) decodeResponse() {
	sb := strings.Builder{}
	param := d.getUInt16(FrameParamMSBOffset)
	switch d.rx[FrameCommandOffset] {
	case ErrorCondition:
		switch param {
		case 0x01:
			sb.WriteString("!! Module busy. Initialization in progress.")
		case 0x02:
			sb.WriteString("!! Module is in Sleep Mode")
		case 0x03:
			sb.WriteString("!! Serial receiving error. Incomplete frame.")
		case 0x04:
			sb.WriteString("!! Incorrect checksum")
		case 0x05:
			sb.WriteString("!! Specified track is out of scope")
		case 0x06:
			sb.WriteString("!! Specified track is not found")
		case 0x07:
			sb.WriteString("!! Insertion error")
		case 0x08:
			sb.WriteString("!! SD card read operation failed")
		case 0x0A:
			sb.WriteString("!! Module entered Sleep Mode")
		default:
			sb.WriteString(fmt.Sprintf("!! Unknown error: %04X", param))
		}
	case 0x41:
		sb.WriteString("!! Module feedback. Unsupported")
	case 0x43:
		// Query volume
	case 0x44:
		// Query EQ
	case 0x47:
		// Query number of tracks in the root of USB flash dri
	case 0x48:
		// Query number of tracks in the root of micro SD card
	case 0x4B:
		// Query current track in the USB flash drive
	case 0x4C:
		// Query current track in the micro SD Card
	case 0x4E:
		// Query number of tracks in a folder
	case 0x4F:
		// Query number of folders in the current storage device
	case GetStatus:
		msb := uint8(param >> 8)
		lsb := uint8(param & 0xFF)
		switch msb {
		case 0x01:
			sb.WriteString(">> USB track")
		case 0x02:
			sb.WriteString(">> SD track")
		case 0x10:
			sb.WriteString(">> Module in sleep mode")
		default:
			sb.WriteString(fmt.Sprintf(" -> unexpected 'device' MSB param : %02X", msb))
		}
		if msb != 0x10 {
			switch lsb {
			case 0x00:
				sb.WriteString(" stopped")
			case 0x01:
				sb.WriteString(" playing")
			case 0x02:
				sb.WriteString(" paused")
			case 0x11:
				sb.WriteString("")
				// Unknown code. Appears in the course of a track playing at regular intervals
			default:
				sb.WriteString(fmt.Sprintf(" -> unexpected LSB param: %02X", lsb))
			}
		} else {
			sb.WriteString("")
		}
	case 0x3A:
		switch param {
		case 0x01:
			sb.WriteString(">> USB flash drive is plugged in")
		case 0x02:
			sb.WriteString(">> SD card is plugged in")
		case 0x04:
			sb.WriteString(">> USB cable connected to PC is plugged in")
		}
	case 0x3B:
		switch param {
		case 0x01:
			sb.WriteString(">> USB flash drive removed")
		case 0x02:
			sb.WriteString(">> SD card removed")
		case 0x04:
			sb.WriteString(">> USB cable disconnected from PC")
		}
	case 0x03C:
		sb.WriteString(">> USB track finished playing")
	case 0x3D:
		sb.WriteString(">> SD card track finished playing")
	case 0x3F:
		switch param {
		case 0x01:
			sb.WriteString(">> USB flash drive online")
		case 0x02:
			sb.WriteString(">> SD card online")
		case 0x03:
			sb.WriteString(">> PC online")
		case 0x04:
			sb.WriteString(">> USB flash drive + SD card online")
		}
	default:
		sb.WriteString(fmt.Sprintf(">> Unexpected command: 0x%02X", d.rx[FrameCommandOffset]))
	}

	if d.debug >= 2 || (d.rx[FrameCommandOffset] == ErrorCondition && d.debug >= 1) {
		println(sb.String())
	}
}
