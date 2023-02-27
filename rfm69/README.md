### Credits!

This is port of the driver here https://github.com/charles-d-burton/rfm69-1, which is based on https://github.com/chbmuc/rfm69, 
which in turn is based on https://github.com/fulr/rfm69

### The port

This port only contains the radio handling part as I think the higher level part for routing packets from 
here https://github.com/charles-d-burton/rfm69-1/blob/master/handler.go should be done in it's own package

### NOTE: This is only tested with RPI Pico, AVR boards will fail for sure as there is no ISR setup for them in the machine package.