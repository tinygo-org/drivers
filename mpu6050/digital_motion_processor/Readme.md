# MPU6050 Digital Motion Processor (DMP) device driver for Tinygo

A mpu6050 gyro that just works. Thanks to these smart people for smoothing out the
wrinkles. 

Inspired by Maker's Wharf which is a C/C++ implemntation of Jeff Rowberg's MPU6050
Arduino library. See it here: 
[https://www.youtube.com/watch?v=k5i-vE5rZR0](https://www.youtube.com/watch?v=k5i-vE5rZR0)  

Jeff Rowberg's Arduino library : 
[https://github.com/ElectronicCats/mpu6050](https://github.com/ElectronicCats/mpu6050)  


## Implementation

Tested with the arduino-zero(SAMD21) Tinygo machine.   

To run:  clone the repository, `go mod install` or `go get tinygo.org/x/drivers`   

Then `tinygo flash --target=arduino-zero -monitor main.go`.  

It will flash the arduino zero (or similar samd21 based) board with the connected MPU6050 on the default i2c ports.
( checkout the video for hardware connection to the MPU6050).   

Use the tinygo usb driver - it is really just a matter of identifying and then connecting See video.  

Note: The demo does not use interrupts  

## Usage

When the Arduino/MPU6050 starts up it will run the calibration for all 6-axis'.
It should be placed on a flat horizontal surface and kept STILL. Once calibration 
has completed, the yaw, pitch and roll will be dumped at approx 100 millisecond intervals.   

The output can be used together with VPython for simple animation.  

### VPython animation

The repo has simple bash scripts to get VPython up and running. 
A python virtual environment is the easiest to get going.  
Use the scripts as a guide:  
  - create the Vitual env.  
  - activate the venv `source ./bin/activate`  
  - install the vpython and pyserial modules with pip3  
  - edit the Animation.py to reference your usb serial connection. Use `python3 listUSBports` to find your usb connection.  
  - run `python3 Animation.py` to view the DMP output.



