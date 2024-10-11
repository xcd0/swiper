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
	led             machine.Pin // 基板上のLED
	mutex           sync.Mutex
	pwm_ch          uint8
	pwm_for_monitor PWM      // モニター用正弦波出力用PWM
	calced          []uint32 // 正弦波のルックアップテーブルから計算したもの
)

func main() {
	ch := make(chan ePushState, 1) // キー入力用channel 先行入力できるほど人間はつよくないので容量1。
	q := make(chan struct{})       // 処理完了フラグ用channel
	buf := make([]ePushState, 0, 20)
	// メインの処理。
	{
		go debug()
		go OutputSignal(&s, ch, q, &buf) // 出力信号作成スレッド    : channelから受け取り、信号を生成して出力する。
		LoopPinCheck(&s, ch, q, &buf)    // キー入力監視(メイン)スレッド: ピン状態を読み取り、channelに投げる。
	}
}
