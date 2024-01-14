package sd

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[statusRsvd0-0]
	_ = x[statusRsvd1-1]
	_ = x[statusRsvd2-2]
	_ = x[statusAuthSeqError-3]
	_ = x[statusRsvdSDIO-4]
	_ = x[statusAppCmd-5]
	_ = x[statusFXEvent-6]
	_ = x[statusRsvd7-7]
	_ = x[statusReadyForData-8]
	_ = x[statusEraseReset-13]
	_ = x[statusECCDisabled-14]
	_ = x[statusWPEraseSkip-15]
	_ = x[statusCSDOverwrite-16]
	_ = x[statusGenericError-19]
	_ = x[statusControllerError-20]
	_ = x[statusECCFailed-21]
	_ = x[statusIllegalCommand-22]
	_ = x[statusComCRCError-23]
	_ = x[statusLockUnlockFailed-24]
	_ = x[statusCardIsLocked-25]
	_ = x[statusWPViolation-26]
	_ = x[statusEraseParamError-27]
	_ = x[statusEraseSeqError-28]
	_ = x[statusBlockLenError-29]
	_ = x[statusAddrError-30]
	_ = x[statusAddrOutOfRange-31]
}

const (
	_status_name_0 = "Rsvd0Rsvd1Rsvd2AuthSeqErrorRsvdSDIOAppCmdFXEventRsvd7ReadyForData"
	_status_name_1 = "EraseResetECCDisabledWPEraseSkipCSDOverwrite"
	_status_name_2 = "GenericErrorControllerErrorECCFailedIllegalCommandComCRCErrorLockUnlockFailedCardIsLockedWPViolationEraseParamErrorEraseSeqErrorBlockLenErrorAddrErrorAddrOutOfRange"
)

var (
	_status_index_0 = [...]uint8{0, 5, 10, 15, 27, 35, 41, 48, 53, 65}
	_status_index_1 = [...]uint8{0, 10, 21, 32, 44}
	_status_index_2 = [...]uint8{0, 12, 27, 36, 50, 61, 77, 89, 100, 115, 128, 141, 150, 164}
)

func (i status) string() string {
	switch {
	case i <= 8:
		return _status_name_0[_status_index_0[i]:_status_index_0[i+1]]
	case 13 <= i && i <= 16:
		i -= 13
		return _status_name_1[_status_index_1[i]:_status_index_1[i+1]]
	case 19 <= i && i <= 31:
		i -= 19
		return _status_name_2[_status_index_2[i]:_status_index_2[i+1]]
	default:
		return ""
	}
}

func (s status) String() string {
	return string(s.appendf(nil, ','))
}

func (s status) appendf(b []byte, delim byte) []byte {
	b = append(b, '[')
	if s == 0 {
		return append(b, ']')
	}
	for bit := 0; bit < 32; bit++ {
		if s&(1<<bit) != 0 {
			b = append(b, status(bit).string()...)
			b = append(b, delim)
		}
	}
	b = append(b[:len(b)-1], ']')
	return b
}
