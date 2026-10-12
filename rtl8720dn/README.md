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

## Host tests

Run the RPC and socket tests with Go on the host.

```sh
go test -race -tags rtl8720dn_hosttest ./rtl8720dn
```

The test tag excludes hardware setup and uses a test device with an
`io.ReadWriter` UART. The tests use the driver RPC, `Connect()`, and `Close()`
methods. `make unit-test` also runs these tests.
