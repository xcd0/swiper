package main

import (
	"device/rp"
	"log"
	"machine"
	"runtime"
	"time"
)

var (
	gpio []machine.Pin
	led  machine.Pin // 基板上のLED
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

	for {
		if !led.Get() {
			led.High()
		} else {
			led.Low()
		}
		time.Sleep(time.Second)
		//time.Sleep(time.Millisecond * 500)
		debug()
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
