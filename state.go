package main

import (
	"fmt"
	"time"
)

type ePushState int

const (
	PUSH_NONE          ePushState = 0         //   0 何も押されていない
	PUSH_DIT                      = 1 << iota //   1 短音が押されている
	PUSH_DASH                                 //   2 長音が押されている
	PUSH_STRAIGHT                             //   4 ストレートキー用ピンが押されている。
	PUSH_ADD_SPEED                            //   8 スピードアップが押されている
	PUSH_SUB_SPEED                            //  16 スピードダウンが押されている
	PUSH_ADD_FREQUENCY                        //  32 周波数ダウンピンが押されている
	PUSH_SUB_FREQUENCY                        //  64 周波数ダウンピンが押されている
	PUSH_ADD_DEBOUNCE                         // 128 デバウンスアップピンが押されている
	PUSH_SUB_DEBOUNCE                         // 256 デバウンスダウンピンが押されている
	PUSH_REVERSE                              // 512 パドルの長短切り替えピンが押されている
)

type PushState struct {
	Now ePushState // 現在の状態

	SpeedOffset    int  // 初期値を20WPMとして、オフセットを保持する。初期値0。
	FreqOffset     int  // 初期値を440Hzとして、オフセットを保持する。初期値0。
	DebounceOffset int  // 初期値を20msとして、オフセットを保持する。初期値0。
	Reverse        bool // 長短のパドル反転。

	tick     time.Duration // 1つの短音の長さ(ms)。SpeedOffsetから計算する。SpeedOffsetが0の時20WPMになるように計算する。
	freq     int           // モニター用ビープ音の周波数。FreqOffsetから計算する。初期値700Hz。
	debounce time.Duration // チャタリング防止のための待機時間(ms)。DebounceOffsetから計算する。初期値20ms。
	harf     time.Duration // 半周期分の時間(us)
}

func (s *PushState) Update() {
	//s.Pre= s.NowState
	s.Now = PUSH_NONE
	if pin_dit.Get() {
		s.Now |= PUSH_DIT
	}
	if pin_dash.Get() {
		s.Now |= PUSH_DASH
	}
	if pin_straight.Get() {
		s.Now |= PUSH_STRAIGHT
	}
	if pin_add_speed.Get() {
		s.Now |= PUSH_ADD_SPEED
	}
	if pin_sub_speed.Get() {
		s.Now |= PUSH_SUB_SPEED
	}
	if pin_add_frequency.Get() {
		s.Now |= PUSH_ADD_FREQUENCY
	}
	if pin_sub_frequency.Get() {
		s.Now |= PUSH_SUB_FREQUENCY
	}
	if pin_add_debounce.Get() {
		s.Now |= PUSH_ADD_DEBOUNCE
	}
	if pin_sub_debounce.Get() {
		s.Now |= PUSH_SUB_DEBOUNCE
	}
	if pin_reverse.Get() {
		s.Now |= PUSH_REVERSE
	}
	{ // スピード
		s.SpeedOffset = Clamp(s.SpeedOffset, -20, 10) // とりあえず5wpmから35wpmまでとする。
		wpm := (s.SpeedOffset + 25)                   // とりあえず初期値を25wpmとする。wpmから1つの短音の長さを計算する。
		// 1wpmは1分間にPARIS(50短点)を1回送る速さ。 例えば24wpmの短点は50ms、長点は150msになる。
		// つまり、n[wpm]は、1分間に(n*50)短点(1秒間にn*50/60短点)の速さなので、1短点は60/50/n*1000[ms]の長さになる。
		s.tick = time.Duration(1000*60/50/wpm) * time.Millisecond
	}

	{ // 音程
		s.FreqOffset = Clamp(s.FreqOffset, -8, 8) // とりあえず、800を基準に、400から1200までを狙って50Hz刻みとする。
		s.freq = (s.FreqOffset*50 + 800)          // とりあえず、初期値を800Hzとする。
	}
	{ // デバウンス
		s.DebounceOffset = Clamp(s.DebounceOffset, -2, 18) // 20msを基準に0msから200msまでを狙って10ms刻みとする。
		s.debounce = time.Duration(s.DebounceOffset*10 + 20)
	}
	{
		// 周波数がs.freqになるようにいい感じにピンを切り替える。
		// 半周期分の時間(us)
		s.harf = time.Duration(1000*1000/2/s.freq) * time.Microsecond
	}
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

func (e ePushState) String() string {
	switch e {
	case PUSH_NONE:
		return "NONE"
	case PUSH_DIT:
		return "dit"
	case PUSH_DASH:
		return "dash"
	case PUSH_ADD_SPEED:
		return "sp_up"
	case PUSH_SUB_SPEED:
		return "sp_dn"
	case PUSH_ADD_FREQUENCY:
		return "fq_up"
	case PUSH_SUB_FREQUENCY:
		return "fq_dn"
	case PUSH_ADD_DEBOUNCE:
		return "de_up"
	case PUSH_SUB_DEBOUNCE:
		return "de_dn"
	case PUSH_REVERSE:
		return "reverse"
	}
	return ""
}

func (ps *PushState) String() string {
	return fmt.Sprintf("State:%v, wpm:%v, t:%v, f:%v, h:%v, d:%v, so:%v, fo:%v, do:%v, rv:%v",
		ps.Now,
		Clamp(ps.SpeedOffset, -15, 16)+20,
		ps.tick,
		ps.freq,
		ps.harf,
		ps.debounce,
		ps.SpeedOffset,
		ps.FreqOffset,
		ps.DebounceOffset,
		ps.Reverse,
	)
}
