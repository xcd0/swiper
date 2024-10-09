//go:generate go run generate_sin_table.go

package main

import (
	"log"
	"machine"
)

func init() {
	log.SetFlags(log.Ltime)

	initialize_gpio()
	initialize()
}

func initialize_gpio() {
	gpio = []machine.Pin{
		machine.GPIO0, machine.GPIO1, machine.GPIO2, machine.GPIO3, machine.GPIO4, machine.GPIO5, machine.GPIO6, machine.GPIO7, machine.GPIO8, machine.GPIO9,
		machine.GPIO10, machine.GPIO11, machine.GPIO12, machine.GPIO13, machine.GPIO14, machine.GPIO15, machine.GPIO16, machine.GPIO17, machine.GPIO18, machine.GPIO19,
		machine.GPIO20, machine.GPIO21, machine.GPIO22, machine.GPIO23, machine.GPIO24, machine.GPIO25, machine.GPIO26, machine.GPIO27, machine.GPIO28, machine.GPIO29,
	}
	{
		// ピン設定。
		// 使用できるモード。
		/*
			machine.PinOutput
			machine.PinInput
			machine.PinInputPulldown
			machine.PinInputPullup
			machine.PinAnalog
			machine.PinUART
			machine.PinPWM
			machine.PinI2C
			machine.PinSPI
			machine.PinPIO0
			machine.PinPIO1
		*/
		for c := range gpio {
			gpio[c].Configure(
				machine.PinConfig{
					//Mode: machine.PinOutput,
					//Mode: machine.PinInput,
					//Mode: machine.PinInputPulldown,
					//Mode: machine.PinInputPullup,
					Mode: machine.PinInputPulldown,
				},
			)
		}
	}
	{
		// 基板上のLEDの設定。
		led = machine.LED
		led.Configure(machine.PinConfig{Mode: machine.PinOutput})
		led.Low()
	}
}

func initialize() {
	// 初期設定。
	init_flash() // flashの設定。設定値の読み書きに使用する。
	{
		savedSetting, err := readSettingFromFile(filesystem, setting_filepath)
		if err != nil {
			log.Printf("main: %v. Using default settings.\r\n", err)
			s.setting = NewSetting()
		} else {
			s.setting = savedSetting
		}
		s.preSetting = s.setting
	}
	{
		// 設定からGPIOの状態を更新する。
		// 入出力GPIOポート設定。
		// 出力ピン設定。
		pin_out.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin_beep_out.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	{
		if p := s.setting.PinSetting.OutputSine; !s.setting.UseAnalogOutput {
			if 26 <= p || p <= 29 {
				log.Printf("main: init_pwm")
				// アナログ出力設定。
				// 未実装。
			} else {
				log.Printf("main: Analog output is specified, but no analog output pin is assigned.") // アナログ出力ピンが指定されていない。
				// PWM設定。
				init_pwm(p)
			}
		} else {
			// PWM設定。
			init_pwm(p)
		}
	}
}
