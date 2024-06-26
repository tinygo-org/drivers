package backlight

type Driver interface {
	SetBrightness(uint8)
}
