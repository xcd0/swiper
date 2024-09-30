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
		led = machine.LED
		led.Configure(machine.PinConfig{Mode: machine.PinOutput})
		led.Low()
	}
	{
		pin_out = gpio[0]           // 出力ピン
		pin_beep_out = gpio[1]      // モニター用サウンド出力ピン
		pin_straight = gpio[2]      // 単式電鍵用ピン
		pin_dit = gpio[3]           // 複式電鍵用短音ピン
		pin_dash = gpio[4]          // 複式電鍵用長音ピン
		pin_add_speed = gpio[5]     // スピードアップピン
		pin_sub_speed = gpio[6]     // スピードダウンピン
		pin_add_frequency = gpio[7] // 周波数アップピン
		pin_sub_frequency = gpio[8] // 周波数ダウンピン
		pin_add_debounce = gpio[9]  // デバウンスアップピン
		pin_sub_debounce = gpio[10] // デバウンスダウンピン
		pin_reverse = gpio[11]      // 長短反転ピン

		// 出力ピン設定。
		pin_out.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin_beep_out.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
}
