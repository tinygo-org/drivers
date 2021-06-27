# RTL8720DN Driver

This package provides a driver to use a separate connected WiFi processor `RTL8720DN` for TCP/UDP communication.
At this time, only part of TCP is supported.

## Using th RTL8720DN Driver

For now, it is only available for the `RTL8720DN` on `Wio Terminal`.
You can try the following command.

```
$ tinygo flash --target wioterminal --size short ./examples/rtl8720dn/webclient/
$ tinygo flash --target wioterminal --size short ./examples/rtl8720dn/tlsclient/
```

## RTL8720DN Firmware

Follow the steps below to update.
The firmware must be version 2.1.2 or later.

https://wiki.seeedstudio.com/Wio-Terminal-Network-Overview/