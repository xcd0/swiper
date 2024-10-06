//go:generate go run generate_sin_table.go

package main

import (
	"log"
	"machine"
)

func init() {
	log.SetFlags(log.Ltime)
	gpio = []machine.Pin{
		machine.GPIO0, machine.GPIO1, machine.GPIO2, machine.GPIO3, machine.GPIO4, machine.GPIO5, machine.GPIO6, machine.GPIO7, machine.GPIO8, machine.GPIO9,
		machine.GPIO10, machine.GPIO11, machine.GPIO12, machine.GPIO13, machine.GPIO14, machine.GPIO15, machine.GPIO16, machine.GPIO17, machine.GPIO18, machine.GPIO19,
		machine.GPIO20, machine.GPIO21, machine.GPIO22, machine.GPIO23, machine.GPIO24, machine.GPIO25, machine.GPIO26, machine.GPIO27, machine.GPIO28, machine.GPIO29,
	}
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
	{
		// 基板上のLEDの設定。
		led = machine.LED
		led.Configure(machine.PinConfig{Mode: machine.PinOutput})
		led.Low()
	}
	{
		// 入出力GPIOポート設定。
		i := 0
		SetGPIO := func(pinnum *int) machine.Pin {
			(*pinnum)++
			return gpio[(*pinnum)-1]
		}
		pin_out = SetGPIO(&i)           //  0 出力ピン
		pin_beep_out = SetGPIO(&i)      //  1 モニター用サウンド出力ピン
		pin_straight = SetGPIO(&i)      //  2 単式電鍵用ピン
		pin_dit = SetGPIO(&i)           //  3 複式電鍵用短音ピン
		pin_dash = SetGPIO(&i)          //  4 複式電鍵用長音ピン
		pin_reset = SetGPIO(&i)         //  5 設定リセットピン
		pin_add_speed = SetGPIO(&i)     //  6 スピードアップピン
		pin_sub_speed = SetGPIO(&i)     //  7 スピードダウンピン
		pin_add_frequency = SetGPIO(&i) //  8 周波数アップピン
		pin_sub_frequency = SetGPIO(&i) //  9 周波数ダウンピン
		pin_add_debounce = SetGPIO(&i)  // 10 デバウンスアップピン
		pin_sub_debounce = SetGPIO(&i)  // 11 デバウンスダウンピン
		pin_reverse = SetGPIO(&i)       // 12 長短反転ピン

		// 出力ピン設定。
		pin_out.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin_beep_out.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}

	{
		pwm := machine.PWM0 // GPIO1でPWMする場合PWM0を指定すればよい。
		pwm.Configure(machine.PWMConfig{Period: uint64(5e2)})
		calced = make([]uint32, len(sinTable))
		for i := 0; i < len(sinTable); i++ {
			calced[i] = uint32(float32(pwm.Top()) * sinTable[i])
		}
	}
	{
		log.Printf("init: init_flash")
		// flashの設定。設定値の読み書きに使用する。
		init_flash()
	}
}
