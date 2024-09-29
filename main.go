package main

import (
	"fmt"
	"machine"
	"time"
)

var (
	gpio              []machine.Pin
	led               machine.Pin // 基板上のLED
	pin_out           machine.Pin // 出力ピン
	pin_beep_out      machine.Pin // モニター用サウンド出力ピン
	pin_dit           machine.Pin // 短音ピン
	pin_dash          machine.Pin // 長音ピン
	pin_add_speed     machine.Pin // スピードアップピン
	pin_sub_speed     machine.Pin // スピードダウンピン
	pin_add_frequency machine.Pin // 周波数アップピン
	pin_sub_frequency machine.Pin // 周波数ダウンピン
	pin_add_debounce  machine.Pin // デバウンスアップピン
	pin_sub_debounce  machine.Pin // デバウンスダウンピン
	pin_reverse       machine.Pin // パドルの長短切り替えピン
)

func main() {
	if false {
		// デバッグ用。
		go func() {
			for {
				time.Sleep(time.Millisecond * 100)
				led.Low()
				time.Sleep(time.Millisecond * 100)
				led.High()
			}
		}()
	}

	var s PushState
	ch := make(chan ePushState, 1) // キー入力用channel 先行入力できるほど人間はつよくないので容量1。
	q := make(chan struct{})       // 処理完了フラグ用channel

	go OutputSignal(&s, ch, q) // 出力信号作成スレッド    : channelから受け取り、信号を生成して出力する。
	LoopPinCheck(&s, ch, q)    // キー入力監視(メイン)スレッド: ピン状態を読み取り、channelに投げる。
}

func OutputSignal(s *PushState, ch chan ePushState, q chan struct{}) {
	for {
		select {
		case ps := <-ch: // ePushStateが入ってくる。
			if ps == PUSH_DIT {
				if !s.Reverse {
					fmt.Printf(".")
					Output(s, 1)
				} else {
					fmt.Printf("-")
					Output(s, 3)
				}
			} else if ps == PUSH_DASH {
				if !s.Reverse {
					fmt.Printf("-")
					Output(s, 3)
				} else {
					fmt.Printf(".")
					Output(s, 1)
				}
			} else if ps == PUSH_ADD_SPEED {
				s.SpeedOffset++
			} else if ps == PUSH_SUB_SPEED {
				s.SpeedOffset--
			} else if ps == PUSH_ADD_FREQUENCY {
				s.FreqOffset++
			} else if ps == PUSH_SUB_FREQUENCY {
				s.FreqOffset--
			} else if ps == PUSH_ADD_DEBOUNCE {
				s.DebounceOffset++
			} else if ps == PUSH_SUB_DEBOUNCE {
				s.DebounceOffset--
			} else if ps == PUSH_REVERSE {
				s.Reverse = !s.Reverse
			} else {
				//
			}

			// メインスレッドの待機を終了する。
			q <- struct{}{}
		default:
			//log.Println("OutputSignal: default")
			time.Sleep(time.Millisecond)
			// 何もしない
		}
	}
}

func Output(s *PushState, ticks int) {
	pin_out.High()
	led.High()
	{
		// t[ms]の間ビープを生成する。
		end := time.Now().Add(s.tick * time.Duration(ticks)) // 終了時刻を計算
		//log.Printf("ht: %vus", us)
		for time.Now().Before(end) {
			time.Sleep(s.harf)
			pin_beep_out.High()
			time.Sleep(s.harf)
			pin_beep_out.Low()
		}
	}
	pin_out.Low()
	led.Low()
	// 文字ごとの間隔 3tick空ける。
	time.Sleep(s.tick * time.Duration(3))
	// 単語の間は4tick空ける必要があるが、このプログラムでは単語の判断は無理なのでユーザーが頑張るものとする。
}

func LoopPinCheck(s *PushState, ch chan ePushState, q chan struct{}) {

	for {
		time.Sleep(time.Millisecond * time.Duration(1))
		preState := s.Now
		s.Update()
		if s.Now == PUSH_NONE {
			continue
		}

		// 同時押しは無視する?
		// 長押しは無視する?

		// チャタリング防止のためデバウンス期間待つ。
		{
			f := s.Now != preState // ひとつ前の状態がfalseでこのループでtrueになったかを調べる。
			if !f {
				continue
			}
			//log.Printf("sleep for chattering: %vms", s.debounce)
			time.Sleep(s.debounce)
			// 再度チェック。
			s.Update()
			f = s.Now != preState && s.Now != PUSH_NONE
			if !f {
				continue
			}
		}

		// 再度チェックしてtrueだったので出力する。
		ch <- s.Now

		// 別スレッドの処理が終わるまで待つ。
		<-q

		// 長押しでリピートする機能。
		if false {
			// ここに来た場合信号出力が終わっている。
			// その時点で長押しされている場合複数回リピートしたい。
			// そのためここでリセットしておく。
			s.Now = PUSH_NONE
		}
	}
}
