# RTL8720DN Driver

This package provides a driver to use a separate connected WiFi processor `RTL8720DN` for TCP/UDP communication.

## Using th RTL8720DN Driver

For now, it is only available for the `RTL8720DN` on `Wio Terminal`.
You can try the following command.

```
$ tinygo flash --target wioterminal --size short ./examples/net/webclient/
$ tinygo flash --target wioterminal --size short ./examples/net/tlsclient/
```

## RTL8720DN Firmware

Follow the steps below to update.
The firmware must be version 2.1.2 or later.

https://wiki.seeedstudio.com/Wio-Terminal-Network-Overview/
