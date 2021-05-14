# SPI sdcard/mmc driver

This package provides the driver for sdcard/mmc with SPI connection.  
`examples/sdcard/console` shows a low-level access example.  
`examples/sdcard/tinyfs` shows an example of using fatfs to read FAT32.  

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
