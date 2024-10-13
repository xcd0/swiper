package main

import (
	"machine"
	"sync"
	"time"
)

const (
	char_dit      = "." // 短音
	char_dash     = "-" // 長音
	char_space    = " " // 文字区切り
	char_straight = "*" // 任意長さストレートキー

	long_press = time.Duration(200) * time.Millisecond // 長押し扱いする秒数
)

var (
	s               PushState
	gpio            []machine.Pin
	led             machine.Pin      // 基板上のLED
	rb              *InputRingBuffer // 1文字分のモールス信号を保持するリングバッファ。
	mutex_rb        sync.Mutex
	mutex_sine      sync.Mutex
	pwm_ch          uint8
	pwm_for_monitor PWM      // モニター用正弦波出力用PWM
	calced          []uint32 // 正弦波のルックアップテーブルから計算したもの
)

func main() {

	sig_ch := make(chan struct{})    // キー入力用channel
	quit_ch := make(chan struct{})   // 処理完了フラグ用channel
	rb = NewInputInputRingBuffer(20) // 1文字分のモールス信号を保持するリングバッファ。

	// メインの処理。
	{
		go debug()
		go OutputSignal(sig_ch, quit_ch) // 出力信号作成スレッド        : 信号を生成して出力する。
		LoopPinCheck(sig_ch, quit_ch)    // キー入力監視(メイン)スレッド: ピン状態を読み取る。
	}
}
