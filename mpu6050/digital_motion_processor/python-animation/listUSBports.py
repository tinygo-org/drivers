from serial.tools import list_ports
port = list(list_ports.comports())
for p in port:
    print(p.device)
