# ST7735 driver

There are multiple devices using the ST7735 chip, and there are multiple versions ST7735B, ST7735R & ST7735S. Two apparently identical displays might have different configurations. The most common issues are:

* Colors are inverted (black is white and viceversa), invert the colors with display.InvertColors(true)
* Colors are not right (red is blue and viceversa, but green is ok), some displays uses BRG instead of RGB for defining colors, change the mode with display.IsBGR(true)
* There is noise/snow/confetti in the screen, probably rows and columns offsets are wrong, configure them with st7735.Config{RowOffset:XX, ColumnOffset:YY}

If nothing of the above works, your device may need a different boot-up process.