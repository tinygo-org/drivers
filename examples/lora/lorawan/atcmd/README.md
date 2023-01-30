# AT-CMD implementation of at-lora command set

This example implements the AT command set as used by Seeed in the LoRa-E5 series of boards, but in the form of a TinyGo program that provides a serial interface. 

See https://files.seeedstudio.com/products/317990687/res/LoRa-E5%20AT%20Command%20Specification_V1.0%20.pdf for more information.

```
$ tinygo monitor
Connected to /dev/ttyACM0. Press Ctrl-C to exit.
+AT: OK
+VER: 0.0.1 (sx127x v18)
```

# Building

Run the following commands from the main `drivers` directory.

## Simulator

Builds/flashes atcmd console application with simulator instead of actual LoRa radio.

```
tinygo flash -target pico ./examples/lora/lorawan/atcmd/
```

## PyBadge with LoRa Featherwing 

Builds/flashes atcmd console application on PyBadge using LoRa Featherwing (RFM95/SX1276).

```
tinygo flash -target pybadge -tags featherwing ./examples/lora/lorawan/atcmd/
```

## LoRa-E5 

Builds/flashes atcmd console application on Lora-E5 using onboard SX126x.

```
tinygo flash -target lorae5 ./examples/lora/lorawan/atcmd/
```

## Joining a Public Lorawan Network

```
AT+ID=DevEui,0101010101010101
AT+ID=AppEui,0123012301230213
AT+KEY=APPKEY,AEAEAEAEAEAEAEAAEAEAEAEAEAEAAEAE
AT+LW=NET,ON  
AT+JOIN
```

AT+LW=NET,(ON|OFF) command changes Lora Sync Word to connect on public network(ON) or private networks(OFF)
