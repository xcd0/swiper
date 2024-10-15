package main

import (
	"machine"
	"time"

	_ "tinygo.org/x/drivers/ssd1306"
)

func init_oled() {
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.I2C0.Configure(machine.I2CConfig{})
	err := lcdInit()
	if err != nil {
	}
	time.Sleep(time.Millisecond * 200)
	for {
		lcdString("Hello")
		time.Sleep(time.Millisecond * 500)
	}
}

func lcdInit() error {
	cmds := []byte{0x38, 0x39, 0x14, 0x73, 0x56, 0x6C, 0x38, 0x01, 0x0C}

	for i := 0; i < len(cmds); i++ {
		err := machine.I2C0.Tx(0x7c>>1, []byte{0x00, cmds[i]}, nil)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}

func lcdString(str string) {
	for i := 0; i < len(str); i++ {
		machine.I2C0.Tx(0x7c>>1, []byte{0x40, str[i]}, nil)
		time.Sleep(time.Millisecond * 1)
	}
}
