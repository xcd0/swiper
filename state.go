package main

import (
	"fmt"
	"time"
)

type ePushState int

const (
	PUSH_NONE          ePushState = 0         //    0 何も押されていない
	PUSH_DIT                      = 1 << iota //    1 短音が押されている
	PUSH_DASH                                 //    2 長音が押されている
	PUSH_STRAIGHT                             //    4 ストレートキー用ピンが押されている。
	PUSH_RESET                                //    8 パドルの長短切り替えピンが押されている
	PUSH_ADD_SPEED                            //   16 スピードアップが押されている
	PUSH_SUB_SPEED                            //   32 スピードダウンが押されている
	PUSH_ADD_FREQUENCY                        //   64 周波数ダウンピンが押されている
	PUSH_SUB_FREQUENCY                        //  128 周波数ダウンピンが押されている
	PUSH_ADD_DEBOUNCE                         //  256 デバウンスアップピンが押されている
	PUSH_SUB_DEBOUNCE                         //  512 デバウンスダウンピンが押されている
	PUSH_REVERSE                              // 1024 パドルの長短切り替えピンが押されている
)

// 外部から変更可能な設定値。変更時にflashに保存する。起動時に読み込みを試みる。
type Setting struct {
	ID             uint32 // これは データの有効性を判定するためのマジックナンバー。flashメモリ上にあるデータが正しく初期化されているかの判定に使う。
	SpeedOffset    int    // 初期値を20WPMとして、オフセットを保持する。初期値0。
	FreqOffset     int    // 初期値を440Hzとして、オフセットを保持する。初期値0。
	DebounceOffset int    // 初期値を20msとして、オフセットを保持する。初期値0。
	Reverse        bool   // 長短のパドル反転。
}

type PushState struct {
	Now ePushState // 現在の状態

	setting Setting // 外部から変更可能な設定値。変更時にflashに保存する。起動時に読み込みを試みる。

	tick     time.Duration // 1つの短音の長さ(ms)。SpeedOffsetから計算する。SpeedOffsetが0の時20WPMになるように計算する。
	freq     int           // モニター用ビープ音の周波数。FreqOffsetから計算する。初期値700Hz。
	debounce time.Duration // チャタリング防止のための待機時間(ms)。DebounceOffsetから計算する。初期値20ms。
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
	if pin_reset.Get() {
		s.Now |= PUSH_RESET
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
	case PUSH_STRAIGHT:
		return "straight"
	case PUSH_RESET:
		return "reset"
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
	return fmt.Sprintf("State:%v, wpm:%v, t:%v, f:%v, d:%v, so:%v, fo:%v, do:%v, rv:%v",
		ps.Now,
		Clamp(ps.setting.SpeedOffset, -15, 16)+20,
		ps.tick,
		ps.freq,
		ps.debounce,
		ps.setting.SpeedOffset,
		ps.setting.FreqOffset,
		ps.setting.DebounceOffset,
		ps.setting.Reverse,
	)
}
