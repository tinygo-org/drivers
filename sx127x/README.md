# SX127x LoRa Radio

This processor comes in several different packages.

## Adafruit LoRa Radio Featherwing

https://www.adafruit.com/product/3231

This is the LoRa 9x @ 900 MHz radio version, which can be used for either 868MHz or 915MHz transmission/reception - the exact radio frequency is determined when you load the software since it can be tuned around dynamically. They can easily go 2 Km line of sight using simple wire antennas, or up to 20Km with directional antennas and settings tweaks.

### Connecting

To use the LoRa Featherwing with a Pybadge/Gobadge, you need to connect some pads on the board itself. Solder some jumper wires as follows:

RST  -> A
CS   -> B
DIO1 -> C
IRQ  -> D 

You also need to connect an antenna. The easiest is to cut a small piece of wire to the correct length based on the desired frequency:

EU868 Mhz - 34.54 cm

You can also use a fancier solution such as a uFL SMA connector with matching antenna.
