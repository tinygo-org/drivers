module tinygo.org/x/drivers

go 1.15

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/frankban/quicktest v1.10.2
	github.com/sago35/tinygo-dma v0.0.0-20210610020721-297675ab9b23 // indirect
	github.com/tinygo-org/tinyfs v0.0.0-20210514090915-924e60a7bcf8
)

replace (
    github.com/sago35/tinygo-dma => ../../sago35/tinygo-dma
)
