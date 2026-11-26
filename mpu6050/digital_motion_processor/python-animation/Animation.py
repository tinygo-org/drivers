# Check this youtube video to get
# details of this work 
# https://www.youtube.com/watch?v=k5i-vE5rZR0
from vpython import *
import math
import serial

def rotationInfo( ringPos, arrowOffset, angleName):
    """
    Returns ring with arrow indicating direction of rotation along a given axis 

    ringPos:        ring position
    arrowOffset:    arrow offset from ring center
    angleName:      this will be one of 'yaw, pitch or roll'
    """

    # Ring
    rotRing = ring(pos = ringPos, axis = ringPos, radius = arrowOffset.mag,
                   thickness = 0.04)
    # Arrow
    ringArrow = arrow(pos = ringPos + arrowOffset, axis = cross(ringPos, arrowOffset),
                      radius = 0.2, thickness = 0.6, length = 0.3, shaftwidth = 0.05)
    # TextLabel
    ringText = text(pos = ringPos * (1 + (0.4 * ringPos.mag)), text = angleName,
                    color = color.white, height = 0.2, depth = 0.05, align = 'center',
                    up = vector(0, 0, -1))
    return compound([rotRing, ringArrow, ringText])

def setScene():
    """
    Set scene properties and creates objects that are fixed
    """

    # Scene
    scene.range = 5
    # scene.forward = vector( -0.8, -1.2, -0.8) # uncomment to get a more 3D view
    scene.background = color.cyan
    scene.width = 1200
    scene.height = 1000

    title = text(pos=vec(0, 3, 0), text='MPU6050', align='center', color=color.blue, height = 0.4, depth = 0.2)

    return title

def createRotatingObjects():
    # MPU6050 module
    # Original orientation as per video
    mpu =  box(length = 4, height = 2, width = .2, opacity = 0.3, pos =  vector(0, 0, 0), color = color.blue )
    # orientation as per my board
    # mpu =  box(length = 2, height = .02, width = 4, opacity = 0.3, pos =  vector(0, 0, 0), color = color.blue )

    # markings on MPU6050 module
    Yaxis = arrow(length = 2, shaftwidth = 0.1, axis = mpu.axis, color = color.white)
    Xaxis = arrow(length = 0.7, shaftwidth = 0.1, axis = vector(0, 0, 1), color = color.white)
    XaxisLabel = text(pos = vector(-0.6, 0.1, 0.7), text = 'X-axis', color = color.white,
                      height = 0.2, depth = 0.05, align = 'center', up = vector(0, 0, -1))
    YaxisLabel = text(pos = vector(1, 0.1, -0.1), text = 'Y-axis', color = color.white,
                      height = 0.2, depth = 0.05, align = 'center', up = vector(0, 0, -1))
    topSideLabel = text(pos = vector(-1.2, 0.1, -0.2), text = 'Top side', color = color.white,
                      height = 0.2, depth = 0.05, align = 'center', up = vector(0, 0, -1))

    # Rings with arrows indicating yaw, pitch and roll direction of rotation
    rotInfoY = rotationInfo(vector(2.4, 0, 0), vector( 0, 0, 0.2), 'roll')
    rotInfoX = rotationInfo(vector(0, 0, 1.3), vector( 0.2, 0, 0), 'pitch')
    rotInfoZ = rotationInfo(vector(0, -1, 0), vector( -0.2, 0, 0), 'yaw')
   
    return compound( [mpu, Xaxis, Yaxis, XaxisLabel, YaxisLabel, topSideLabel, 
                    rotInfoY, rotInfoX, rotInfoZ])

def rodriquesRotation(v, k, angle):
    return v * cos(angle) + cross(k, v) * sin(angle)

title = setScene()
angles = label(pos=vec(-2.5, 2, 1), text='yaw= 0\npitch= 0\nroll= 0\n', align='center', color=color.black, height = 30, depth = 0)
rotatingObjects = createRotatingObjects()

# initialize serial port
ArduinoSerial = serial.Serial('/dev/cu.usbmodem14401', 115200)


while (True):
    # data might not be available at first so keep on trying
    try:
        # grab angles from arduino
        while(ArduinoSerial.inWaiting() == 0):
            pass
        dataPacket = ArduinoSerial.readline()
        dataPacket = str(dataPacket, 'utf-8')
        splitPacket = dataPacket.split("\t")
        yaw = float(splitPacket[0])
        pitch = float(splitPacket[1])
        roll = float(splitPacket[2])

        # update display based on rotation
        rate(50) # vpython needs this, it is sort of display update rate.

        # display values while they are still in degrees
        angles.text = f'yaw: {round(yaw)}\npitch: {round(pitch)}\n roll: {round(roll)}'
        
        yaw = math.radians(yaw)
        pitch = math.radians(pitch)
        roll = math.radians(roll)

        # Let x and y point along the MPU's X and Y axes respectively. Let's stipulate
        # that at zero yaw, pitch and roll - x and y are given by:
        x = vector(0, 0, 1)
        y = vector(1, 0, 0)

        # Let's apply the yaw rotation to x and y using the rodriques formula:
        x = rodriquesRotation(x, vector(0, -1, 0), yaw)
        y = rodriquesRotation(y, vector(0, -1, 0), yaw)

        # note x is invariant  under pitch rotaion. Apply the pitch rotation to y;
        y = rodriquesRotation(y, x, pitch)

        # y is invariant to the roll rotation. Apply the roll rotation to x:
        x = rodriquesRotation(x, y, roll)

        rotatingObjects.axis = y
        # draw the result  by updating the up vector
        rotatingObjects.up = cross(x, y)
    except:
        pass

