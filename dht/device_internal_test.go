package dht

import (
	"testing"
)

func TestDeviceType_extractData(t *testing.T) {
	bitStr := "0000001010001100000000010101111111101110"
	buf := bitStringToBytes(bitStr)

	tt := []struct {
		name     string
		d        DeviceType
		buf      []byte
		wantTemp int16
		wantHum  uint16
	}{
		{
			// temp = 35.1C hum = 65.2%
			name: "DHT22", d: DHT22, buf: buf, wantTemp: 351, wantHum: 652,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotTemp, gotHum := tc.d.extractData(tc.buf)
			if gotTemp != tc.wantTemp {
				t.Errorf("extractData() gotTemp = %v, want %v", gotTemp, tc.wantTemp)
			}
			if gotHum != tc.wantHum {
				t.Errorf("extractData() gotHum = %v, want %v", gotHum, tc.wantHum)
			}
		})
	}
}

func bitStringToBytes(s string) []byte {
	b := make([]byte, (len(s)+(8-1))/8)
	for i, r := range s {
		if r < '0' || r > '1' {
			panic("not in range")
		}
		b[i>>3] |= byte(r-'0') << uint(7-i&7)
	}
	return b
}
