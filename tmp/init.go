//go:generate go run generate_sin_table.go

package main

import (
	"log"
	"machine"
	"time"
)

func init() {
	log.SetFlags(log.Ltime)
	if _debug {
		time.Sleep(time.Second * 1)
		log.Printf("debug: wait 1s")
	}
	gpio = []machine.Pin{
		machine.GPIO0, machine.GPIO1, machine.GPIO2, machine.GPIO3, machine.GPIO4, machine.GPIO5, machine.GPIO6, machine.GPIO7, machine.GPIO8, machine.GPIO9,
		machine.GPIO10, machine.GPIO11, machine.GPIO12, machine.GPIO13, machine.GPIO14, machine.GPIO15, machine.GPIO16, machine.GPIO17, machine.GPIO18, machine.GPIO19,
		machine.GPIO20, machine.GPIO21, machine.GPIO22, machine.GPIO23, machine.GPIO24, machine.GPIO25, machine.GPIO26, machine.GPIO27, machine.GPIO28, machine.GPIO29,
	}
	// すべて一旦Pulldown。正論理でも負論理でもよいが統一すること。ここではとりあえず正論理にしている。
	for c := range gpio {
		gpio[c].Configure(
			machine.PinConfig{
				Mode: machine.PinInputPulldown,
			},
		)
	}
}
