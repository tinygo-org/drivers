package rtl8720dn

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/net"
)

type Driver struct {
	*RTL8720DN
}

// New returns a new RTL8720DN driver. The UART that is passed in
// will be reconfigured at the baud rate required by the device.
func New(uart *machine.UART, tx, rx, en machine.Pin) *Driver {
	enable(en)
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx, BaudRate: 614400})

	return &Driver{
		RTL8720DN: &RTL8720DN{
			port:  &UARTx{UART: uart},
			seq:   1,
			sema:  make(chan bool, 1),
			debug: false,
		},
	}
}

func (d *Driver) Configure() error {
	net.UseDriver(d)

	_, err := d.Rpc_tcpip_adapter_init()
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) ConnectToAccessPoint(ssid, pass string, timeout time.Duration) error {
	if len(ssid) == 0 {
		return net.ErrWiFiMissingSSID
	}

	return d.ConnectToAP(ssid, pass)
}

func (d *Driver) Disconnect() error {
	_, err := d.Rpc_wifi_disconnect()
	return err
}

func (d *Driver) GetClientIP() (string, error) {
	ip, _, _, err := d.GetIP()
	return ip.String(), err
}
