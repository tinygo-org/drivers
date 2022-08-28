package rtl8720dn

import (
	"fmt"
	"time"
)

func (d *Driver) ConnectToAP(ssid string, password string) error {
	if len(ssid) == 0 || len(password) == 0 {
		return fmt.Errorf("connection failed: either ssid or password not set")
	}

	_, err := d.Rpc_wifi_off()
	if err != nil {
		return err
	}
	_, err = d.Rpc_wifi_on(0x00000001)
	if err != nil {
		return err
	}

	_, err = d.Rpc_wifi_disconnect()
	if err != nil {
		return err
	}

	numTry := 5
	securityType := uint32(0x00400004)
	for i := 0; i < numTry; i++ {
		ret, err := d.Rpc_wifi_connect(ssid, password, securityType, -1, 0)
		if err != nil {
			return err
		}
		if ret != 0 {
			if i == numTry-1 {
				return fmt.Errorf("connection failed: rpc_wifi_connect failed")
			}
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}

	_, err = d.Rpc_tcpip_adapter_dhcpc_start(0)
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		_, err = d.Rpc_wifi_is_connected_to_ap()
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (d *Driver) GetIP() (ip, subnet, gateway IPAddress, err error) {
	ip_info := make([]byte, 12)
	_, err = d.Rpc_tcpip_adapter_get_ip_info(0, &ip_info)
	if err != nil {
		return nil, nil, nil, err
	}

	ip = IPAddress(ip_info[0:4])
	subnet = IPAddress(ip_info[4:8])
	gateway = IPAddress(ip_info[8:12])

	return ip, subnet, gateway, nil
}
