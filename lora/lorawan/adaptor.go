package lorawan

import "tinygo.org/x/drivers/lora"

var ActiveRadio lora.Radio

func UseRadio(r lora.Radio) {
	if ActiveRadio != nil {
		panic("lorawan.ActiveRadio is already set")
	}
	ActiveRadio = r
}

func Join() error {
	return nil
}

func SendUplink() error {
	return nil
}

func ListenDownlink() error {
	return nil
}
