# AT-CMD implementation of at-lora command set

```
$ tinygo monitor
Connected to /dev/ttyACM0. Press Ctrl-C to exit.
+AT: OK
+VER: 0.0.1 (sx127x v18)
```

# Building

## Simulator

```
tinygo flash -target pico ./examples/lora/lorawan/atcmd/
```

## PyBadge with LoRa Featherwing 

```
tinygo flash -target pybadge -tags featherwing ./examples/lora/lorawan/atcmd/
```

## LoRa-E5 

```
tinygo flash -target lorae5 ./examples/lora/lorawan/atcmd/
```

