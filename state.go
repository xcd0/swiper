package main

import (
	"fmt"
	"math"
	"time"
)

type PinState uint32

type PushState struct {
	Now PinState // 現在のGPIOの状態

	setting    Setting // 外部から変更可能な設定値。変更時にflashに保存する。起動時に読み込みを試みる。
	preSetting Setting // 直前の設定状態。設定をflashに書き込むかどうかの判定に使う。できるだけflashに書き込みたくないので。

	dit  time.Duration // 1つの短音の長さ。
	dash time.Duration // 1つの長音の長さ。
}

func ReadGPIO() PinState {
	var state PinState = 0
	for i, p := range gpio {
		if false || //
			i == 23 || i == 24 || i == 25 || i == 29 || //
			i == s.setting.PinSetting.I2CSDA || //
			i == s.setting.PinSetting.I2CSCL || //
			i == s.setting.PinSetting.Output || //
			i == s.setting.PinSetting.AnalogChangeSpeed || //
			i == s.setting.PinSetting.AnalogChangeFrequency || //
			false {
			// 入力ピン以外は無視する。
		} else {
			if p.Get() {
				state |= 1 << i
			}
		}
	}
	return state
}

func DebugPrintGPIO() string {
	reverse := func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}
	// GPIO0からGPIO29までを左から右に
	//fmt.Printf("gpio: %v\n", reverse(fmt.Sprintf("%029b", s.Now)))
	return fmt.Sprintf("gpio: %v\n", reverse(fmt.Sprintf("%029b", s.Now)))
}

func AreAllPinsOff(t PinState) bool {
	// 出力ピンはReadGPIO()で0になるのでこれでよい。
	return t == 0
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
