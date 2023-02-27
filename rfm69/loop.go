package rfm69

import (
	"fmt"
)

// Send data
func (r *Device) Send(d *Data) {
	r.tx <- d
}

func (r *Device) loop() {
	println("In loop")
	err := r.SetMode(RF_OPMODE_RECEIVER)
	if err != nil {
		fmt.Print(err)
	}
	defer r.SetMode(RF_OPMODE_STANDBY)

	for {
		select {
		case dataToTransmit := <-r.tx:
			println("in send")
			// TODO: can send?
			r.readWriteReg(REG_PACKETCONFIG2, 0xFB, RF_PACKET2_RXRESTART) // avoid RX deadlocks
			err = r.SetModeAndWait(RF_OPMODE_STANDBY)
			if err != nil {
				fmt.Print(err)
			}
			err = r.writeReg(REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_00)
			if err != nil {
				fmt.Print(err)
			}
			println("sending ..")
			err = r.writeFifo(dataToTransmit)
			if err != nil {
				fmt.Print(err)
			}
			println("changing to transmit mode...")
			err = r.SetMode(RF_OPMODE_TRANSMITTER)
			if err != nil {
				fmt.Print(err)
			}
			println("Waiting (block) for irq chan...")
			<-r.irq
			println("Changing back to receive")
			err = r.SetModeAndWait(RF_OPMODE_STANDBY)
			if err != nil {
				fmt.Print(err)

			}
			err = r.writeReg(REG_DIOMAPPING1, RF_DIOMAPPING1_DIO0_01)
			if err != nil {
				fmt.Print(err)

			}
			err = r.SetMode(RF_OPMODE_RECEIVER)
			if err != nil {
				fmt.Print(err)

			}
		case interrupt := <-r.irq:
			if interrupt {
				println("triggered intr")
				if r.mode != RF_OPMODE_RECEIVER {
					continue
				}
				flags, err := r.readReg(REG_IRQFLAGS2)
				if err != nil {
					fmt.Print(err)

				}
				if flags&RF_IRQFLAGS2_PAYLOADREADY == 0 {
					continue
				}
				data, err := r.readFifo()
				if err != nil {
					fmt.Print(err)

				}
				println("Trigering OnReceive")
				if r.Config.OnReceive != nil {
					go r.Config.OnReceive(&data)
				}
				err = r.SetMode(RF_OPMODE_RECEIVER)
				if err != nil {
					fmt.Print(err)
				}
			}
		case <-r.quit:
			r.quit <- true
			return
		}
	}
}
