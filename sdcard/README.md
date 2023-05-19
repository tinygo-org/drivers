# SPI sdcard/mmc driver

This package provides the driver for sdcard/mmc with SPI connection.

To use a file system on the SDcard, please see the TinyFS repo:

https://github.com/tinygo-org/tinyfs

See `examples/sdcard/console` for a low-level access example.

## Stack size

If you use this package, you need to set `default-stack-size` in `targets/*.json`.  
For example, `targets/wioterminal.json` has the following configuration.  

```
{
    "inherits": ["atsamd51p19a"],
    "build-tags": ["wioterminal"],
    "flash-1200-bps-reset": "true",
    "flash-method": "msd",
    "msd-volume-name": "Arduino",
    "msd-firmware-name": "firmware.uf2",
    "default-stack-size": 2048
}
```
