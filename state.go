package main

import (
	"math"
	"time"
)

type PinState uint32

type PushState struct {
	Now PinState // 現在のGPIOの状態

	setting    Setting // 外部から変更可能な設定値。変更時にflashに保存する。起動時に読み込みを試みる。
	preSetting Setting // 直前の設定状態。設定をflashに書き込むかどうかの判定に使う。できるだけflashに書き込みたくないので。

	tick     time.Duration // 1つの短音の長さ(ms)。SpeedOffsetから計算する。SpeedOffsetが0の時20WPMになるように計算する。
	freq     int           // モニター用ビープ音の周波数。FreqOffsetから計算する。初期値700Hz。
	debounce time.Duration // チャタリング防止のための待機時間(ms)。DebounceOffsetから計算する。初期値20ms。

	recorded_input []string
}

func ReadGPIO() PinState {
	var state PinState = 0
	for i, p := range gpio {
		if i < 23 && 25 < i || i == s.setting.PinSetting.Output || i == s.setting.PinSetting.AnalogChangeSpeed || i == s.setting.PinSetting.AnalogChangeFrequency {
			// 入力ピン以外は無視する。
		} else {
			if p.Get() {
				state |= 1 << i
			}
		}
	}
	return state
}

func clamp(f, l, h float64) float64 {
	return math.Min(h, math.Max(l, f))
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
