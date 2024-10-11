package main

import (
	"fmt"
	"machine"
)

// machineのPWMは非公開なので。
type PWM interface {
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
	Enable(bool)
	Top() uint32
	Configure(config machine.PWMConfig) error
	Channel(machine.Pin) (uint8, error)
}

// PWMの設定。
func init_pwm(pin int) {
	var pwm_frequency uint64 = 5e2

	var err error
	pwm_for_monitor, pwm_ch, err = GetPWM(pin, pwm_frequency)
	if err != nil {
		handleError(fmt.Errorf("init_pwm: %v", err))
		return
	}
	calced = make([]uint32, len(sinTable))
	for i := 0; i < len(sinTable); i++ {
		calced[i] = uint32(float32(pwm_for_monitor.Top()) * sinTable[i])
	}
}

func GetPWM(pin int, frequency uint64) (PWM, uint8, error) {
	gpio := []machine.Pin{
		machine.GPIO0, machine.GPIO1, machine.GPIO2, machine.GPIO3, machine.GPIO4, machine.GPIO5, machine.GPIO6, machine.GPIO7, machine.GPIO8, machine.GPIO9,
		machine.GPIO10, machine.GPIO11, machine.GPIO12, machine.GPIO13, machine.GPIO14, machine.GPIO15, machine.GPIO16, machine.GPIO17, machine.GPIO18, machine.GPIO19,
		machine.GPIO20, machine.GPIO21, machine.GPIO22, machine.GPIO23, machine.GPIO24, machine.GPIO25, machine.GPIO26, machine.GPIO27, machine.GPIO28, machine.GPIO29,
	}
	return getPWM(gpio[pin], frequency)
}

func getPWM(pin machine.Pin, frequency uint64) (PWM, uint8, error) {
	pwms := [...]PWM{machine.PWM0, machine.PWM1, machine.PWM2, machine.PWM3, machine.PWM4, machine.PWM5, machine.PWM6, machine.PWM7}
	slice, err := machine.PWMPeripheral(pin)
	if err != nil {
		return nil, 0, err
	}
	pwm := pwms[slice]
	err = pwm.Configure(machine.PWMConfig{Period: 1e9 / frequency}) // 100Hz for starters.
	if err != nil {
		return nil, 0, err
	}
	channel, err := pwm.Channel(pin)
	return pwm, channel, err
}
