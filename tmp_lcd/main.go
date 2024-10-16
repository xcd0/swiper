//go:generate go run generate_sin_table.go
package main

import (
	"device/rp"
	"fmt"
	"image/color"
	"log"
	"machine"
	"math"
	"time"

	"tinygo.org/x/drivers/ssd1306"

	font "github.com/Nondzu/ssd1306_font"
)

var (
	gpio            []machine.Pin
	led             machine.Pin // 基板上のLED
	display         ssd1306.Device
	pwm_for_monitor PWM // モニター用正弦波出力用PWM。
	pwm_ch          uint8
	calced          []uint32             // 正弦波のルックアップテーブルから計算したもの。
	fc              uint64   = 100 * 1e3 // 搬送波の周波数
	f               int      = 700       // 出力する正弦波の周波数
)

func init() {
	log.SetFlags(log.Ltime)
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

	time.Sleep(time.Millisecond * 100)
	I2CSDA := 20                                                    // I2C0SDA
	I2CSCL := 21                                                    // I2C0SCL
	gpio[I2CSDA].Configure(machine.PinConfig{Mode: machine.PinI2C}) // I2C0
	gpio[I2CSCL].Configure(machine.PinConfig{Mode: machine.PinI2C}) // I2C0
	//machine.I2C0.Configure(machine.I2CConfig{Frequency: machine.TWI_FREQ_400KHZ})
	machine.I2C0.Configure(machine.I2CConfig{})

	time.Sleep(time.Second)
}

func main() {
	log.Printf("main: start")
	defer main_end()

	// debug_blink()
	debug_lcd()

	for i := 0; i < 50; i++ {
		if !led.Get() {
			led.High()
		} else {
			led.Low()
		}
		time.Sleep(time.Millisecond * time.Duration(100))
		output_info()
	}

	debug_pwm()
}

func debug_lcd() {
	for i := 0; i < 10; i++ {
		log.Printf("debug_lcd: start ")
		time.Sleep(time.Second * 1)
	}
	defer log.Printf("debug_lcd: start ")
	for {
		log.Printf("lcd: line")
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

		log.Printf("lcd: font")
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
	}
}

func debug_pwm() {
	log.Printf("debug_pwm")

	//fc = 200 * 1e3 // 搬送波の周波数
	f = 700 // 出力する正弦波の周波数
	rate := 1000.

	// led
	{
		// 出力する正弦波の1波長を何分割して出力するかを計算。
		pick := int((1. / float64(f)) / (1. / float64(fc)))                    // 分割数
		step := Clamp(len(sinTable)/pick, 1, len(sinTable))                    // 正弦波を1000分割したlookuptableを用意しているので、これをpick数で分割した1stepを計算する。
		stepDuration := time.Second / time.Duration(int(float64(f*pick)*rate)) // 1ステップの期間。

		for {
			log.Printf("pick: %v, step: %v, stepDuration: %v", pick, step, stepDuration)
			pwm_for_monitor = InitPwm(25)
			for j := 0; j < f; j++ {
				for i := 0; i < len(sinTable); i += step {
					pwm_for_monitor.Set(pwm_ch, calced[i])
					time.Sleep(stepDuration)
				}
			}
			// end := time.Now().Add(stepDuration * time.Duration(f))
			// led = machine.LED
			// led.Configure(machine.PinConfig{Mode: machine.PinOutput})
			// j := 0
			// for time.Now().Before(end) {
			// 	if !led.Get() {
			// 		led.High()
			// 	} else {
			// 		led.Low()
			// 	}
			// 	time.Sleep(time.Millisecond * time.Duration(100))
			// 	if j%10000 == 0 {
			// 		//log.Printf("pwm: non pwm led loop end j: %v, duration: %v", j, stepDuration*time.Duration(f))
			// 	}
			// 	j++
			// }
			//log.Printf("pwm: non pwm led loop end j: %v, duration: %v", j, stepDuration*time.Duration(f))
			log.Printf("pwm: non pwm led loop end: duration: %v", stepDuration*time.Duration(f))
			log.Printf("pwm: loop end")
		}
	}
}

func InitPwm(pwmpinnum int) PWM {
	pwm_pin := gpio[pwmpinnum]
	pwm_pin.Configure(machine.PinConfig{Mode: machine.PinPWM})
	slice, err := machine.PWMPeripheral(pwm_pin)
	if err != nil {
		handleError(fmt.Errorf("init_pwm: %v", err))
	}
	log.Printf("machine.PWMPeripheral(%v): %v", pwm_pin, slice)
	pwms := [...]PWM{machine.PWM0, machine.PWM1, machine.PWM2, machine.PWM3, machine.PWM4, machine.PWM5, machine.PWM6, machine.PWM7}
	pwm_for_monitor = pwms[slice]
	if err = pwm_for_monitor.Configure(machine.PWMConfig{Period: 1e9 / fc}); err != nil {
		handleError(fmt.Errorf("init_pwm: %v", err))
	}
	if pwm_ch, err = pwm_for_monitor.Channel(pwm_pin); err != nil {
		handleError(fmt.Errorf("init_pwm: %v", err))
	}
	calced = make([]uint32, len(sinTable))
	for i := 0; i < len(sinTable); i++ {
		calced[i] = uint32(float32(pwm_for_monitor.Top()) * sinTable[i])
	}
	log.Printf("pwm: initialized.")
	return pwm_for_monitor
}

// エラー時の無限ループ処理
func handleError(err error) {
	for {
		fmt.Printf("%v\r\n", err)
		time.Sleep(time.Second)
	}
}

// machineのPWMは非公開なのでinterfaceを定義。
type PWM interface {
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
	Enable(bool)
	Top() uint32
	Configure(config machine.PWMConfig) error
	Channel(machine.Pin) (uint8, error)
}

func output_info() {
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
		IMPLEMENTER  : %08x
		VARIANT      : %08x
		ARCHITECTURE : %08x
		PARTNO       : %08x
		REVISION     : %08x
	CHIP_ID
		REVISION     : %08x
		PART         : %08x
		MANUFACTURER : %08x`,
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

func debug_blink() {
	for {
		if !led.Get() {
			led.High()
		} else {
			led.Low()
		}
		time.Sleep(time.Second)
		output_info()
	}
}

func main_end() {
	for {
		log.Printf("main: end.")
		time.Sleep(time.Millisecond * 1000)
	}
}

func clamp(f, l, h float32) float32 {
	return float32(math.Min(float64(h), math.Max(float64(l), float64(f))))
}

func Clamp(f, low, high int) int {
	if f < low {
		return low
	}
	if f > high {
		return high
	}
	return f
}
