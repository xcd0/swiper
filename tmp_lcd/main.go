package main

import (
	"device/rp"
	"image/color"
	"log"
	"machine"
	"runtime"
	"time"

	"tinygo.org/x/drivers/ssd1306"

	font "github.com/Nondzu/ssd1306_font"
)

var (
	gpio    []machine.Pin
	led     machine.Pin // 基板上のLED
	display ssd1306.Device
)

func init() {
	gpio = []machine.Pin{
		machine.GPIO0, machine.GPIO1, machine.GPIO2, machine.GPIO3, machine.GPIO4, machine.GPIO5, machine.GPIO6, machine.GPIO7, machine.GPIO8, machine.GPIO9,
		machine.GPIO10, machine.GPIO11, machine.GPIO12, machine.GPIO13, machine.GPIO14, machine.GPIO15, machine.GPIO16, machine.GPIO17, machine.GPIO18, machine.GPIO19,
		machine.GPIO20, machine.GPIO21, machine.GPIO22, machine.GPIO23, machine.GPIO24, machine.GPIO25, machine.GPIO26, machine.GPIO27, machine.GPIO28, machine.GPIO29,
	}
	for c := range gpio {
		gpio[c].Configure(machine.PinConfig{Mode: machine.PinInputPulldown}) // PinOutput, PinInput, PinInputPulldown, PinInputPullup, PinAnalog, PinUART, PinPWM, PinI2C, PinSPI, PinPIO0, PinPIO1
	}
	led = machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Low()

	runtime.Gosched()
}

func main() {
	log.SetFlags(log.Ltime)

	// for {
	// 	if !led.Get() {
	// 		led.High()
	// 	} else {
	// 		led.Low()
	// 	}
	// 	time.Sleep(time.Second)
	// 	debug()
	// }

	I2CSDA := 20 // I2C0SDA
	I2CSCL := 21 // I2C0SCL

	gpio[I2CSDA].Configure(machine.PinConfig{Mode: machine.PinI2C}) // I2C0
	gpio[I2CSCL].Configure(machine.PinConfig{Mode: machine.PinI2C}) // I2C0

	machine.I2C0.Configure(machine.I2CConfig{Frequency: machine.TWI_FREQ_400KHZ})
	time.Sleep(time.Millisecond * 100)
	for {
		time.Sleep(time.Second * 1)
		{

			// Display
			dev := ssd1306.NewI2C(machine.I2C0)
			dev.Configure(ssd1306.Config{Width: 128, Height: 64, Address: ssd1306.Address_128_32, VccState: ssd1306.SWITCHCAPVCC})
			dev.ClearBuffer()
			dev.ClearDisplay()

			//font library init
			display := font.NewDisplay(dev)
			display.Configure(font.Config{FontType: font.FONT_7x10}) //set font here

			// font.FONT_6x8
			// font.FONT_7x10
			// font.FONT_11x18
			// font.FONT_16x26

			display.YPos = 20                 // set position Y
			display.XPos = 0                  // set position X
			display.PrintText("HELLO WORLD!") // print text

		}
		time.Sleep(time.Second * 1)
		{
			display = ssd1306.NewI2C(machine.I2C0)
			display.Configure(ssd1306.Config{Width: 128, Height: 64, Address: 0x3C, VccState: ssd1306.SWITCHCAPVCC})
			display.ClearDisplay()

			x := int16(0)
			y := int16(0)
			deltaX := int16(1)
			deltaY := int16(1)
			for {
				pixel := display.GetPixel(x, y)
				c := color.RGBA{255, 255, 255, 255}
				if pixel {
					c = color.RGBA{0, 0, 0, 255}
				}
				display.SetPixel(x, y, c)
				display.Display()

				x += deltaX
				y += deltaY

				if x == 0 || x == 127 {
					deltaX = -deltaX
				}

				if y == 0 || y == 63 {
					deltaY = -deltaY
				}
				time.Sleep(1 * time.Millisecond)
			}

		}

	}

	for {

	}
}

func debug() {
	cpuidIMP := rp.PPB.GetCPUID_IMPLEMENTER()
	cpuidVAR := rp.PPB.GetCPUID_VARIANT()
	cpuidARCH := rp.PPB.GetCPUID_ARCHITECTURE()
	cpuidPNO := rp.PPB.GetCPUID_PARTNO()
	cpuidREV := rp.PPB.GetCPUID_REVISION()
	chipREV := rp.SYSINFO.GetCHIP_ID_REVISION()
	chipPART := rp.SYSINFO.GetCHIP_ID_PART()
	chipMANU := rp.SYSINFO.GetCHIP_ID_MANUFACTURER()

	log.Printf(`
	CPUID
	IMPLEMENTER:  %08x
	VARIANT:      %08x
	ARCHITECTURE: %08x
	PARTNO:       %08x
	REVISION:     %08x
	CHIP_ID"
	REVISION:     %08x
	PART:         %08x
	MANUFACTURER: %08x`,
		cpuidIMP,
		cpuidVAR,
		cpuidARCH,
		cpuidPNO,
		cpuidREV,
		chipREV,
		chipPART,
		chipMANU,
	)
}
