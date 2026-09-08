package stusb4500

// Constants defining the important register addresses of the STUSB4500
const (
	REG_BCD_TYPEC_REV_LOW         = 0x06 // BCD_TYPEC_REV_LOW register
	REG_BCD_TYPEC_REV_HIGH        = 0x07 // BCD_TYPEC_REV_HIGH register
	REG_BCD_USBPD_REV_LOW         = 0x08 // BCD_USBPD_REV_LOW register
	REG_BCD_USBPD_REV_HIGH        = 0x09 // BCD_USBPD_REV_HIGH register
	REG_DEVICE_CAPAB_HIGH         = 0x0A // DEVICE_CAPAB_HIGH register
	REG_ALERT_STATUS_1            = 0x0B // ALERT_STATUS_1 register
	REG_ALERT_STATUS_1_MASK       = 0x0C // ALERT_STATUS_1_MASK register
	REG_PORT_STATUS_0             = 0x0D // PORT_STATUS_0 register
	REG_PORT_STATUS_1             = 0x0E // PORT_STATUS_1 register
	REG_TYPEC_MONITORING_STATUS_0 = 0x0F // TYPEC_MONITORING_STATUS_0 register
	REG_TYPEC_MONITORING_STATUS_1 = 0x10 // TYPEC_MONITORING_STATUS_1 register
	REG_CC_STATUS                 = 0x11 // CC_STATUS register
	REG_CC_HW_FAULT_STATUS_0      = 0x12 // CC_HW_FAULT_STATUS_0 register
	REG_CC_HW_FAULT_STATUS_1      = 0x13 // CC_HW_FAULT_STATUS_1 register
	REG_PD_TYPEC_STATUS           = 0x14 // PD_TYPEC_STATUS register
	REG_TYPEC_STATUS              = 0x15 // TYPEC_STATUS register
	REG_PRT_STATUS                = 0x16 // PRT_STATUS register
	REG_PHY_STATUS                = 0x17 // PHY_STATUS register
	REG_CC_CAPABILITY_CTRL        = 0x18 // CC_CAPABILITY_CTRL register
	REG_PRT_TX_CTRL               = 0x19 // PRT_TX_CTRL register
	REG_PD_COMMAND_CTRL           = 0x1A // PD_COMMAND_CTRL register
	REG_MONITORING_CTRL_0         = 0x20 // MONITORING_CTRL_0 register
	REG_MONITORING_CTRL_2         = 0x22 // MONITORING_CTRL_2 register
	REG_RESET_CTRL                = 0x23 // RESET_CTRL register
	REG_VBUS_DISCHARGE_TIME_CTRL  = 0x25 // VBUS_DISCHARGE_TIME_CTRL register
	REG_VBUS_DISCHARGE_CTRL       = 0x26 // VBUS_DISCHARGE_CTRL register
	REG_VBUS_CTRL                 = 0x27 // VBUS_CTRL register
	REG_PE_FSM                    = 0x29 // PE_FSM register
	REG_GPIO_SW_GPIO              = 0x2D // GPIO_SW_GPIO register
	REG_DEVICE_ID                 = 0x2F // DEVICE_ID register
	REG_RX_BYTE_CNT               = 0x30 // RX_BYTE_CNT register
	REG_RX_HEADER_LOW             = 0x31 // RX_HEADER_LOW register
	REG_RX_HEADER_HIGH            = 0x32 // RX_HEADER_HIGH register
	REG_RX_DATA_OBJ1_0            = 0x33 // RX_DATA_OBJ1_0 register
	REG_RX_DATA_OBJ1_1            = 0x34 // RX_DATA_OBJ1_1 register
	REG_RX_DATA_OBJ1_2            = 0x35 // RX_DATA_OBJ1_2 register
	REG_RX_DATA_OBJ1_3            = 0x36 // RX_DATA_OBJ1_3 register
	REG_RX_DATA_OBJ2_0            = 0x37 // RX_DATA_OBJ2_0 register
	REG_RX_DATA_OBJ2_1            = 0x38 // RX_DATA_OBJ2_1 register
	REG_RX_DATA_OBJ2_2            = 0x39 // RX_DATA_OBJ2_2 register
	REG_RX_DATA_OBJ2_3            = 0x3A // RX_DATA_OBJ2_3 register
	REG_RX_DATA_OBJ3_0            = 0x3B // RX_DATA_OBJ3_0 register
	REG_RX_DATA_OBJ3_1            = 0x3C // RX_DATA_OBJ3_1 register
	REG_RX_DATA_OBJ3_2            = 0x3D // RX_DATA_OBJ3_2 register
	REG_RX_DATA_OBJ3_3            = 0x3E // RX_DATA_OBJ3_3 register
	REG_RX_DATA_OBJ4_0            = 0x3F // RX_DATA_OBJ4_0 register
	REG_RX_DATA_OBJ4_1            = 0x40 // RX_DATA_OBJ4_1 register
	REG_RX_DATA_OBJ4_2            = 0x41 // RX_DATA_OBJ4_2 register
	REG_RX_DATA_OBJ4_3            = 0x42 // RX_DATA_OBJ4_3 register
	REG_RX_DATA_OBJ5_0            = 0x43 // RX_DATA_OBJ5_0 register
	REG_RX_DATA_OBJ5_1            = 0x44 // RX_DATA_OBJ5_1 register
	REG_RX_DATA_OBJ5_2            = 0x45 // RX_DATA_OBJ5_2 register
	REG_RX_DATA_OBJ5_3            = 0x46 // RX_DATA_OBJ5_3 register
	REG_RX_DATA_OBJ6_0            = 0x47 // RX_DATA_OBJ6_0 register
	REG_RX_DATA_OBJ6_1            = 0x48 // RX_DATA_OBJ6_1 register
	REG_RX_DATA_OBJ6_2            = 0x49 // RX_DATA_OBJ6_2 register
	REG_RX_DATA_OBJ6_3            = 0x4A // RX_DATA_OBJ6_3 register
	REG_RX_DATA_OBJ7_0            = 0x4B // RX_DATA_OBJ7_0 register
	REG_RX_DATA_OBJ7_1            = 0x4C // RX_DATA_OBJ7_1 register
	REG_RX_DATA_OBJ7_2            = 0x4D // RX_DATA_OBJ7_2 register
	REG_RX_DATA_OBJ7_3            = 0x4E // RX_DATA_OBJ7_3 register
	REG_TX_HEADER_LOW             = 0x51 // TX_HEADER_LOW register
	REG_TX_HEADER_HIGH            = 0x52 // TX_HEADER_HIGH register
	REG_DPM_PDO_NUMB              = 0x70 // DPM_PDO_NUMB register
	REG_DPM_SNK_PDO1_0            = 0x85 // DPM_SNK_PDO1_0 register
	REG_DPM_SNK_PDO1_1            = 0x86 // DPM_SNK_PDO1_1 register
	REG_DPM_SNK_PDO1_2            = 0x87 // DPM_SNK_PDO1_2 register
	REG_DPM_SNK_PDO1_3            = 0x88 // DPM_SNK_PDO1_3 register
	REG_DPM_SNK_PDO2_0            = 0x89 // DPM_SNK_PDO2_0 register
	REG_DPM_SNK_PDO2_1            = 0x8A // DPM_SNK_PDO2_1 register
	REG_DPM_SNK_PDO2_2            = 0x8B // DPM_SNK_PDO2_2 register
	REG_DPM_SNK_PDO2_3            = 0x8C // DPM_SNK_PDO2_3 register
	REG_DPM_SNK_PDO3_0            = 0x8D // DPM_SNK_PDO3_0 register
	REG_DPM_SNK_PDO3_1            = 0x8E // DPM_SNK_PDO3_1 register
	REG_DPM_SNK_PDO3_2            = 0x8F // DPM_SNK_PDO3_2 register
	REG_DPM_SNK_PDO3_3            = 0x90 // DPM_SNK_PDO3_3 register
	REG_RDO_REG_STATUS_0          = 0x91 // RDO_REG_STATUS_0 register
	REG_RDO_REG_STATUS_1          = 0x92 // RDO_REG_STATUS_1 register
	REG_RDO_REG_STATUS_2          = 0x93 // RDO_REG_STATUS_2 register
	REG_RDO_REG_STATUS_3          = 0x94 // RDO_REG_STATUS_3 register
)

const (
	// USB PD policy engine states, as defined in REG_PE_FSM (0x29)
	peSoftReset               = 0x01 // 00000001
	peHardReset               = 0x02 // 00000010
	peSendSoftReset           = 0x03 // 00000011
	peBistCarrierMode         = 0x04 // 00000100
	peSnkStartup              = 0x12 // 00010010
	peSnkDiscovery            = 0x13 // 00010011
	peSnkWaitForCapabilities  = 0x14 // 00010100
	peSnkEvaluateCapabilities = 0x15 // 00010101
	peSnkSelectCapabilities   = 0x16 // 00010110
	peSnkTransitionSink       = 0x17 // 00010111
	peSnkReady                = 0x18 // 00011000
	peSnkReadySending         = 0x19 // 00011001
	peDbCpCheckForVbus        = 0x1A // 00011010
	peErrorrecovery           = 0x30 // 00110000
	peSrcTransitionSupply3    = 0x31 // 00110001
	peSrcTransitionSupply2b   = 0x31 // 00110001
	peSrcGetSinkCap           = 0x31 // 00110001
)

const (
	// Some STUSB4500 devices contain 0x21 in the device ID register (0x2F), and
	// others contain 0x25. See the following discussion:
	//   https://github.com/ardnew/STUSB4500/issues/2
	evalDeviceID = 0x21
	prodDeviceID = 0x25
)

// isDeviceIDValid returns true if and only if the provided id equals one of the
// known STUSB4500 device ID register (0x2F) contents.
func isDeviceIDValid(id uint8) bool {
	return evalDeviceID == id || prodDeviceID == id
}

// statusRegister defines the methods available on a read-only (R/O) status
// register. This includes registers whose contents are cleared when read (R/C).
type statusRegister interface {
	parse(...uint8) error
}

// controlRegister defines the methods available on a read-write (R/W) control
// register. All methods defined on statusRegister are also available on
// controlRegister.
type controlRegister interface {
	parse(...uint8)
	format() []uint8
}

type regStatusSnkPDO struct {
	cons struct {
		operationalCurrent uint32 // 0[10]
		voltage            uint32 // 10[10]
		_                  uint8  // 20[3]
		fastRoleReqCurrent uint8  // 23[2]
		dualRoleData       bool   // 25[1]
		usbCommsCapable    bool   // 26[1]
		unconstrainedPower bool   // 27[1]
		higherCapability   bool   // 28[1]
		dualRolePower      bool   // 29[1]
		fixedSupply        uint8  // 30[2]
	}
	vari struct {
		operatingCurrent uint32 // 0[10]
		minVoltage       uint32 // 10[10]
		maxVoltage       uint32 // 20[10]
		variableSupply   uint8  // 30[2]
	}
	batt struct {
		operatingPower uint32 // 0[10]
		minVoltage     uint32 // 10[10]
		maxVoltage     uint32 // 20[10]
		battery        uint8  // 30[2]
	}
}

func (reg *regStatusSnkPDO) parse(word ...uint8) {
	data := lendU32(word...)
	// fixed supply
	reg.cons.operationalCurrent = (data >> 0) & 0x3FF       // uint32 // 0[10]
	reg.cons.voltage = (data >> 10) & 0x3FF                 // uint32 // 10[10]
	reg.cons.fastRoleReqCurrent = uint8((data >> 23) & 0x3) // uint8  // 23[2]
	reg.cons.dualRoleData = 0 != (data>>25)&0x1             // bool   // 25[1]
	reg.cons.usbCommsCapable = 0 != (data>>26)&0x1          // bool   // 26[1]
	reg.cons.unconstrainedPower = 0 != (data>>27)&0x1       // bool   // 27[1]
	reg.cons.higherCapability = 0 != (data>>28)&0x1         // bool   // 28[1]
	reg.cons.dualRolePower = 0 != (data>>29)&0x1            // bool   // 29[1]
	reg.cons.fixedSupply = uint8((data >> 30) & 0x3)        // uint8  // 30[2]
	// variable supply
	reg.vari.operatingCurrent = (data >> 0) & 0x3FF     // uint32 // 0[10]
	reg.vari.minVoltage = (data >> 10) & 0x3FF          // uint32 // 10[10]
	reg.vari.maxVoltage = (data >> 20) & 0x3FF          // uint32 // 20[10]
	reg.vari.variableSupply = uint8((data >> 30) & 0x3) // uint8  // 30[2]
	// battery
	reg.batt.operatingPower = (data >> 0) & 0x3FF // uint32 // 0[10]
	reg.batt.minVoltage = (data >> 10) & 0x3FF    // uint32 // 10[10]
	reg.batt.maxVoltage = (data >> 20) & 0x3FF    // uint32 // 20[10]
	reg.batt.battery = uint8((data >> 30) & 0x3)  // uint8  // 30[2]
}

// since we don't have unions in Go, each of the sub-structs are formatted as
// three separate 32-bit words at the following positions in the returned slice:
//   cons = [0..3], vari = [4..7], batt = [8..11]
func (reg *regStatusSnkPDO) format() []uint8 {

	var cons, vari, batt uint32

	cons |= (reg.cons.operationalCurrent & 0x3FF) << 0
	cons |= (reg.cons.voltage & 0x3FF) << 10
	cons |= (uint32(reg.cons.fastRoleReqCurrent) & 0x3) << 23
	if reg.cons.dualRoleData {
		cons |= 1 << 25
	}
	if reg.cons.usbCommsCapable {
		cons |= 1 << 26
	}
	if reg.cons.unconstrainedPower {
		cons |= 1 << 27
	}
	if reg.cons.higherCapability {
		cons |= 1 << 28
	}
	if reg.cons.dualRolePower {
		cons |= 1 << 29
	}
	cons |= (uint32(reg.cons.fixedSupply) & 0x3) << 30

	vari |= (reg.vari.operatingCurrent & 0x3FF) << 0
	vari |= (reg.vari.minVoltage & 0x3FF) << 10
	vari |= (reg.vari.maxVoltage & 0x3FF) << 20
	vari |= (uint32(reg.vari.variableSupply) & 0x3) << 30

	batt |= (reg.batt.operatingPower & 0x3FF) << 0
	batt |= (reg.batt.minVoltage & 0x3FF) << 10
	batt |= (reg.batt.maxVoltage & 0x3FF) << 20
	batt |= (uint32(reg.batt.battery) & 0x3) << 30

	data := []uint8{}
	data = append(data, bytes32(cons)...)
	data = append(data, bytes32(vari)...)
	data = append(data, bytes32(batt)...)
	return data
	//	return append(append(append([]uint8{}, bytes32(cons)...), bytes32(vari)...), bytes32(batt)...)
}

type regStatusSrcPDO struct {
	cons struct {
		maxOperatingCurrent uint32 // 0[10]
		voltage             uint32 // 10[10]
		peakCurrent         uint8  // 20[2]
		_                   uint8  // 22[3]
		dataRoleSwap        bool   // 25[1]
		communication       bool   // 26[1]
		externallyPowered   bool   // 27[1]
		suspendSupported    bool   // 28[1]
		dualRolePower       bool   // 29[1]
		fixedSupply         uint8  // 30[2]
	}
	vari struct {
		operatingCurrent uint32 // 0[10]
		minVoltage       uint32 // 10[10]
		maxVoltage       uint32 // 20[10]
		variableSupply   uint8  // 30[2]
	}
	batt struct {
		operatingPower uint32 // 0[10]
		minVoltage     uint32 // 10[10]
		maxVoltage     uint32 // 20[10]
		battery        uint8  // 30[2]
	}
}

func (reg *regStatusSrcPDO) parse(word ...uint8) {
	data := lendU32(word...)
	// fixed supply
	reg.cons.maxOperatingCurrent = (data >> 0) & 0x3FF // uint32 // 0[10]
	reg.cons.voltage = (data >> 10) & 0x3FF            // uint32 // 10[10]
	reg.cons.peakCurrent = uint8((data >> 20) & 0x3)   // uint8  // 20[2]
	reg.cons.dataRoleSwap = 0 != (data>>25)&0x1        // bool   // 25[1]
	reg.cons.communication = 0 != (data>>26)&0x1       // bool   // 26[1]
	reg.cons.externallyPowered = 0 != (data>>27)&0x1   // bool   // 27[1]
	reg.cons.suspendSupported = 0 != (data>>28)&0x1    // bool   // 28[1]
	reg.cons.dualRolePower = 0 != (data>>29)&0x1       // bool   // 29[1]
	reg.cons.fixedSupply = uint8((data >> 30) & 0x3)   // uint8  // 30[2]
	// variable supply
	reg.vari.operatingCurrent = (data >> 0) & 0x3FF     // uint32 // 0[10]
	reg.vari.minVoltage = (data >> 10) & 0x3FF          // uint32 // 10[10]
	reg.vari.maxVoltage = (data >> 20) & 0x3FF          // uint32 // 20[10]
	reg.vari.variableSupply = uint8((data >> 30) & 0x3) // uint8  // 30[2]
	// battery
	reg.batt.operatingPower = (data >> 0) & 0x3FF // uint32 // 0[10]
	reg.batt.minVoltage = (data >> 10) & 0x3FF    // uint32 // 10[10]
	reg.batt.maxVoltage = (data >> 20) & 0x3FF    // uint32 // 20[10]
	reg.batt.battery = uint8((data >> 30) & 0x3)  // uint8  // 30[2]
}

type regStatusRDO struct {
	maxCurrent        uint32 // 0[10]
	operatingCurrent  uint32 // 10[10]
	_                 uint8  // 20[3]
	unchunkMsgSupport bool   // 23[1]
	usbSuspend        bool   // 24[1]
	usbCommsCapable   bool   // 25[1]
	capMismatch       bool   // 26[1]
	giveBack          bool   // 27[1]
	objectPos         uint8  // 28[3]
	_                 uint8  // 31[1]
}

func (reg *regStatusRDO) parse(word ...uint8) {
	data := lendU32(word...)
	reg.maxCurrent = (data >> 0) & 0x3FF
	reg.operatingCurrent = (data >> 10) & 0x3FF
	reg.unchunkMsgSupport = 0 != (data>>23)&0x1
	reg.usbSuspend = 0 != (data>>24)&0x1
	reg.usbCommsCapable = 0 != (data>>25)&0x1
	reg.capMismatch = 0 != (data>>26)&0x1
	reg.giveBack = 0 != (data>>27)&0x1
	reg.objectPos = uint8((data >> 28) & 0x7)
}

type msgUsbpdHeader struct {
	messageType     uint8 // 0[5]
	portDataRole    uint8 // 5[1]
	specRevision    uint8 // 6[2]
	portPowerRole   uint8 // 8[1]
	messageID       uint8 // 9[3]
	dataObjectCount uint8 // 12[3]
	extended        uint8 // 15[1]
}

func (reg *msgUsbpdHeader) parse(word ...uint8) {
	data := lendU16(word...)
	reg.messageType = uint8((data >> 0) & 0x1F)
	reg.portDataRole = uint8((data >> 5) & 0x1)
	reg.specRevision = uint8((data >> 6) & 0x3)
	reg.portPowerRole = uint8((data >> 8) & 0x1)
	reg.messageID = uint8((data >> 9) & 0x7)
	reg.dataObjectCount = uint8((data >> 12) & 0x7)
	reg.extended = uint8((data >> 15) & 0x1)
}

// This register (0x0B, access=R/C) indicates an Alert has occurred.
type regStatusAlert struct {
	phy       bool  // 0[1]
	prt       bool  // 1[1]
	_         uint8 // 2[1]
	typeC     bool  // 3[1]
	hwFault   bool  // 4[1]
	monitor   bool  // 5[1]
	ccDetect  bool  // 6[1]
	hardReset bool  // 7[1]
}

func (reg *regStatusAlert) parse(word ...uint8) {
	data := lendU8(word...)
	reg.phy = 0 != (data>>0)&0x1
	reg.prt = 0 != (data>>1)&0x1
	reg.typeC = 0 != (data>>3)&0x1
	reg.hwFault = 0 != (data>>4)&0x1
	reg.monitor = 0 != (data>>5)&0x1
	reg.ccDetect = 0 != (data>>6)&0x1
	reg.hardReset = 0 != (data>>7)&0x1
}

type regControlAlert struct {
	phy       bool  // 0[1]
	prt       bool  // 1[1]
	_         uint8 // 2[1]
	typeC     bool  // 3[1]
	hwFault   bool  // 4[1]
	monitor   bool  // 5[1]
	ccDetect  bool  // 6[1]
	hardReset bool  // 7[1]
}

func (reg *regControlAlert) parse(word ...uint8) {
	data := lendU8(word...)
	reg.phy = 0 != (data>>0)&0x1
	reg.prt = 0 != (data>>1)&0x1
	reg.typeC = 0 != (data>>3)&0x1
	reg.hwFault = 0 != (data>>4)&0x1
	reg.monitor = 0 != (data>>5)&0x1
	reg.ccDetect = 0 != (data>>6)&0x1
	reg.hardReset = 0 != (data>>7)&0x1
}

func (reg *regControlAlert) format() []uint8 {
	var data uint8
	if reg.phy {
		data |= 1 << 0
	}
	if reg.prt {
		data |= 1 << 1
	}
	if reg.typeC {
		data |= 1 << 3
	}
	if reg.hwFault {
		data |= 1 << 4
	}
	if reg.monitor {
		data |= 1 << 5
	}
	if reg.ccDetect {
		data |= 1 << 6
	}
	if reg.hardReset {
		data |= 1 << 7
	}
	return []uint8{data}
}

// This register (0x0D, access=R/C) indicates a bit value change has occurred in
// CC_DETECTION_STATUS register (0x0E).
type regStatusCCDetectTrans struct {
	attached bool  // 0[1]
	_        uint8 // 1[7]
}

func (reg *regStatusCCDetectTrans) parse(word ...uint8) {
	data := lendU8(word...)
	reg.attached = 0 != (data>>0)&0x1
}

// This register (0x0E, access=R/O) provides current status of the connection
// detection and corresponding operation modes.
type regStatusCCDetect struct {
	attached         bool  // 0[1]
	vconnSupplyState uint8 // 1[1]
	dataRole         uint8 // 2[1]
	powerRole        uint8 // 3[1]
	startupPowerMode uint8 // 4[1]
	attachMode       uint8 // 5[3]
}

func (reg *regStatusCCDetect) parse(word ...uint8) {
	data := lendU8(word...)
	reg.attached = 0 != (data>>0)&0x1
	reg.vconnSupplyState = (data >> 1) & 0x1
	reg.dataRole = (data >> 2) & 0x1
	reg.powerRole = (data >> 3) & 0x1
	reg.startupPowerMode = (data >> 4) & 0x1
	reg.attachMode = (data >> 5) & 0x7
}

// This register (0x0F, access=R/C) allows to:
//   - Alert about any change that occurs in MONITORING_STATUS (0x10) register
//   - Manage specific USB PD Acknowledge commands
//   - to manage Type-C state machine Acknowledge to USB PD Requests commands
type regStatusTypeCMonTrans struct {
	vconnValid     bool  // 0[1]
	vbusValid      bool  // 1[1]
	vbusVSafe0V    bool  // 2[1]
	vbusReady      bool  // 3[1]
	vbusLowStatus  bool  // 4[1]
	vbusHighStatus bool  // 5[1]
	_              uint8 // 6[2]
}

func (reg *regStatusTypeCMonTrans) parse(word ...uint8) {
	data := lendU8(word...)
	reg.vconnValid = 0 != (data>>0)&0x1
	reg.vbusValid = 0 != (data>>1)&0x1
	reg.vbusVSafe0V = 0 != (data>>2)&0x1
	reg.vbusReady = 0 != (data>>3)&0x1
	reg.vbusLowStatus = 0 != (data>>4)&0x1
	reg.vbusHighStatus = 0 != (data>>5)&0x1
}

// This register (0x10, access=R/O) provides information on current status of
// the VBUS and VCONN voltages monitoring done respectively on VBUS_SENSE input
// pin and VCONN input pin.
type regStatusTypeCMon struct {
	vconnValid    bool  // 0[1]
	vbusValidSink bool  // 1[1]
	vbusVSafe0V   bool  // 2[1]
	vbusReady     bool  // 3[1]
	_             uint8 // 4[4]
}

func (reg *regStatusTypeCMon) parse(word ...uint8) {
	data := lendU8(word...)
	reg.vconnValid = 0 != (data>>0)&0x1
	reg.vbusValidSink = 0 != (data>>1)&0x1
	reg.vbusVSafe0V = 0 != (data>>2)&0x1
	reg.vbusReady = 0 != (data>>3)&0x1
}

type regStatusCC struct {
	cc1State             uint8 // 0[2]
	cc2State             uint8 // 2[2]
	connectResult        bool  // 4[1]
	lookingForConnection bool  // 5[1]
	_                    uint8 // 6[2]
}

func (reg *regStatusCC) parse(word ...uint8) {
	data := lendU8(word...)
	reg.cc1State = (data >> 0) & 0x3
	reg.cc2State = (data >> 2) & 0x3
	reg.connectResult = 0 != (data>>4)&0x1
	reg.lookingForConnection = 0 != (data>>5)&0x1
}

// This register (0x12, access=R/C) indicates a bit value change has occurred in
// HW_FAULT_STATUS (0x13) register. It alerts also when the over-temperature
// condition is met.
type regStatusHWFaultTrans struct {
	vconnSwOVPFault    bool  // 0[1]
	vconnSwOCPFault    bool  // 1[1]
	vconnSwRVPFault    bool  // 2[1]
	vbusVsrcDischFault bool  // 3[1]
	vpuValid           bool  // 4[1]
	vpuOVPFault        bool  // 5[1]
	_                  uint8 // 6[1]
	thermalFault       bool  // 7[1]
}

func (reg *regStatusHWFaultTrans) parse(word ...uint8) {
	data := lendU8(word...)
	reg.vconnSwOVPFault = 0 != (data>>0)&0x1
	reg.vconnSwOCPFault = 0 != (data>>1)&0x1
	reg.vconnSwRVPFault = 0 != (data>>2)&0x1
	reg.vbusVsrcDischFault = 0 != (data>>3)&0x1
	reg.vpuValid = 0 != (data>>4)&0x1
	reg.vpuOVPFault = 0 != (data>>5)&0x1
	reg.thermalFault = 0 != (data>>7)&0x1
}

// This register (0x13, access=R/O) provides information on hardware fault
// conditions related to the internal pull-up voltage in Source power role and
// to the VCONN power switches.
type regStatusHWFault struct {
	vconnSwOVPFault bool  // 0[1]
	vconnSwOCPFault bool  // 1[1]
	vconnSwRVPFault bool  // 2[1]
	vsrcDischFault  bool  // 3[1]
	_               uint8 // 4[1]
	vbusDischFault  bool  // 5[1]
	vpuPresence     bool  // 6[1]
	vpuOVPFault     bool  // 7[1]
}

func (reg *regStatusHWFault) parse(word ...uint8) {
	data := lendU8(word...)
	reg.vconnSwOVPFault = 0 != (data>>0)&0x1
	reg.vconnSwOCPFault = 0 != (data>>1)&0x1
	reg.vconnSwRVPFault = 0 != (data>>2)&0x1
	reg.vsrcDischFault = 0 != (data>>3)&0x1
	reg.vbusDischFault = 0 != (data>>5)&0x1
	reg.vpuPresence = 0 != (data>>6)&0x1
	reg.vpuOVPFault = 0 != (data>>7)&0x1
}

// This register (0x15, access=R/O) provides information on Type-C connection
// status.
type regStatusTypeC struct {
	typeCFSMState uint8 // 0[5]
	pdSnkTxRp     bool  // 5[1]
	pdSrcTxRp     bool  // 6[1]
	reverse       bool  // 7[1]
}

func (reg *regStatusTypeC) parse(word ...uint8) {
	data := lendU8(word...)
	reg.typeCFSMState = (data >> 0) & 0x1F
	reg.pdSnkTxRp = 0 != (data>>5)&0x1
	reg.pdSrcTxRp = 0 != (data>>6)&0x1
	reg.reverse = 0 != (data>>7)&0x1
}

// This register (0x16, access=R/O) provides information on PRT status.
type regStatusPRT struct {
	hardResetReceived bool  // 0[1]
	hardResetDone     bool  // 1[1]
	msgReceived       bool  // 2[1]
	msgSent           bool  // 3[1]
	bistReceived      bool  // 4[1]
	bistSent          bool  // 5[1]
	_                 uint8 // 6[1]
	txError           bool  // 7[1]
}

func (reg *regStatusPRT) parse(word ...uint8) {
	data := lendU8(word...)
	reg.hardResetReceived = 0 != (data>>0)&0x1
	reg.hardResetDone = 0 != (data>>1)&0x1
	reg.msgReceived = 0 != (data>>2)&0x1
	reg.msgSent = 0 != (data>>3)&0x1
	reg.bistReceived = 0 != (data>>4)&0x1
	reg.bistSent = 0 != (data>>5)&0x1
	reg.txError = 0 != (data>>7)&0x1
}

// This register (0x17, access=R/O) provides information on PHY status.
type regStatusPHY struct {
	txMsgFail bool  // 0[1]
	txMsgDisc bool  // 1[1]
	txMsgSucc bool  // 2[1]
	idle      bool  // 3[1]
	_         uint8 // 4[1]
	sopRxType uint8 // 5[3]
}

func (reg *regStatusPHY) parse(word ...uint8) {
	data := lendU8(word...)
	reg.txMsgFail = 0 != (data>>0)&0x1
	reg.txMsgDisc = 0 != (data>>1)&0x1
	reg.txMsgSucc = 0 != (data>>2)&0x1
	reg.idle = 0 != (data>>3)&0x1
	reg.sopRxType = (data >> 5) & 0x7
}

// This register (0x18, access=R/W) allows to change the advertising of the
// current capability and the VCONN supply capability when operating in Source
// power role.
type regControlCCCap struct {
	vconnSupplyEn     bool  // 0[1]
	vconnSwapEn       bool  // 1[1]
	prSwapEn          bool  // 2[1]
	drSwapEn          bool  // 3[1]
	vconnDischargeEn  bool  // 4[1]
	sinkDisconnect    bool  // 5[1]
	currentAdvertised uint8 // 6[2]
}

func (reg *regControlCCCap) parse(word ...uint8) {
	data := lendU8(word...)
	reg.vconnSupplyEn = 0 != (data>>0)&0x1
	reg.vconnSwapEn = 0 != (data>>1)&0x1
	reg.prSwapEn = 0 != (data>>2)&0x1
	reg.drSwapEn = 0 != (data>>3)&0x1
	reg.vconnDischargeEn = 0 != (data>>4)&0x1
	reg.sinkDisconnect = 0 != (data>>5)&0x1
	reg.currentAdvertised = (data >> 6) & 0x3
}

func (reg *regControlCCCap) format() []uint8 {
	var data uint8
	if reg.vconnSupplyEn {
		data |= 1 << 0
	}
	if reg.vconnSwapEn {
		data |= 1 << 1
	}
	if reg.prSwapEn {
		data |= 1 << 2
	}
	if reg.drSwapEn {
		data |= 1 << 3
	}
	if reg.vconnDischargeEn {
		data |= 1 << 4
	}
	if reg.sinkDisconnect {
		data |= 1 << 5
	}
	data |= (reg.currentAdvertised & 0x3) << 6
	return []uint8{data}
}

// This register (0x19, access=R/W) allows PRT TX layer.
type regControlPRTTx struct {
	prtTxSOPMsg      uint8 // 0[3]
	_                uint8 // 3[1]
	prtRetryMsgCount uint8 // 4[2]
	_                uint8 // 6[2]
}

func (reg *regControlPRTTx) parse(word ...uint8) {
	data := lendU8(word...)
	reg.prtTxSOPMsg = (data >> 0) & 0x7
	reg.prtRetryMsgCount = (data >> 4) & 0x3
}

func (reg *regControlPRTTx) format() []uint8 {
	var data uint8
	data |= (reg.prtTxSOPMsg & 0x7) << 0
	data |= (reg.prtRetryMsgCount & 0x3) << 4
	return []uint8{data}
}

// This register (0x1A, access=R/W) allows to send command to PRL or PE internal
// state machines.
type regControlPDCmd struct {
	cmd uint8 // 0[8]
}

const (
	ctrlPDCommand = 0x26
	ctrlSoftReset = 0x0D
)

func (reg *regControlPDCmd) parse(word ...uint8) {
	data := lendU8(word...)
	reg.cmd = data
}

func (reg *regControlPDCmd) format() []uint8 {
	return []uint8{reg.cmd}
}

// This register (0x1D, access=R/W) allows to Reset PHYTX & change device
// Automation level.
type regControlDevice struct {
	_          uint8 // 0[1]
	phyTxReset bool  // 1[1]
	_          uint8 // 2[4]
	pdTopLayer uint8 // 6[2]
}

func (reg *regControlDevice) parse(word ...uint8) {
	data := lendU8(word...)
	reg.phyTxReset = 0 != (data>>1)&0x1
	reg.pdTopLayer = (data >> 6) & 0x3
}

func (reg *regControlDevice) format() []uint8 {
	var data uint8
	if reg.phyTxReset {
		data |= 1 << 1
	}
	data |= (reg.pdTopLayer & 0x3) << 6
	return []uint8{data}
}

// This register (0x23, access=R/W) allows to reset the device by software.
// The SW_RESET_EN bit acts as the hardware RESET pin but it does not command
// the RESET pin.
type regControlReset struct {
	reset bool  // 0[1]
	_     uint8 // 1[7]
}

func (reg *regControlReset) parse(word ...uint8) {
	data := lendU8(word...)
	reg.reset = 0 != (data>>0)&0x1
}

func (reg *regControlReset) format() []uint8 {
	var data uint8
	if reg.reset {
		data |= 1 << 0
	}
	return []uint8{data}
}
