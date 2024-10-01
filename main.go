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
	pin_straight      machine.Pin // ストレートキー用ピン
	pin_add_speed     machine.Pin // スピードアップピン
	pin_sub_speed     machine.Pin // スピードダウンピン
	pin_add_frequency machine.Pin // 周波数アップピン
	pin_sub_frequency machine.Pin // 周波数ダウンピン
	pin_add_debounce  machine.Pin // デバウンスアップピン
	pin_sub_debounce  machine.Pin // デバウンスダウンピン
	pin_reverse       machine.Pin // パドルの長短切り替えピン

	pwm_ch uint8
)

var _wait bool

func main() {
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
			} else {
				//
			}

			if _wait {
				// メインスレッドの待機を終了する。
				q <- struct{}{}
			}
		default:
			//log.Println("OutputSignal: default")
			time.Sleep(time.Millisecond)
			// 何もしない
		}
	}
}

func OutputSine(sineFrequency int, term time.Duration) {
	pwm := machine.PWM0 // GPIO1でPWMする場合PWM0を指定すればよい。
	pwm.Configure(machine.PWMConfig{Period: uint64(5e2)})
	var err error
	if pwm_ch, err = pwm.Channel(pin_beep_out); err != nil {
		println(err.Error())
		return
	}
	calced := make([]uint32, len(sinTable))
	for i := 0; i < len(sinTable); i++ {
		calced[i] = uint32(float32(pwm.Top()) * sinTable[i])
	}
	tick := 10
	steps := len(sinTable) / tick // 正弦波のステップ数
	stepDuration := time.Second / time.Duration(sineFrequency) / time.Duration(steps)

	end := time.Now().Add(term)
	for time.Now().Before(end) {
		for i := 0; i < len(sinTable); i += tick {
			pwm.Set(pwm_ch, calced[i])
			time.Sleep(stepDuration)
		}
	}
}

func Output(s *PushState, ticks int) {
	pin_out.High()
	led.High()
	OutputSine(s.freq, time.Duration(ticks)*s.tick)
	pin_out.Low()
	led.Low()
	// 文字ごとの間隔 3tick空ける。
	time.Sleep(s.tick * time.Duration(3))
	// 単語の間は4tick空ける必要があるが、このプログラムでは単語の判断は無理なのでユーザーが頑張るものとする。
}

func LoopPinCheck(s *PushState, ch chan ePushState, q chan struct{}) {

	for {
		time.Sleep(time.Millisecond * time.Duration(1))
		if CheckChattering(s) {
			continue // チャタリング防止の待機時間内にスイッチがOFFになった。
		}

		// 何れかのピンがONだったのでピンに従って処理する。

		// 設定値変更ピンの処理。
		if ChangeSetting(s) {
			continue // 設定値変更ピンがONになっていた。設定値を変更したので終わる。
		}

		// 設定値変更ピンでない場合は信号を出力する。

		if s.Now == PUSH_STRAIGHT {
			// もしストレートキー用ピンが押されていたら押されている間出力する。つまりチャタリング防止だけしつつそのまま出力する。
			fmt.Printf("*")
			// 押されている間ONにする。
			pin_out.High()
			led.High()
			for {
				// 押されている間beepを鳴らす。
				time.Sleep(s.harf)
				pin_beep_out.High()
				time.Sleep(s.harf)
				pin_beep_out.Low()
				// チェック
				s.Update()
				if s.Now != PUSH_STRAIGHT {
					break
				}
			}
			// OFFになったので終わる。
			pin_out.Low()
			led.Low()
			// 文字ごとの間隔 3tick空ける。
			time.Sleep(s.tick * time.Duration(3))
		} else {
			// ストレートキー用キー以外(長音、単音)は適切に出力する必要がある。

			ch <- s.Now

			if _wait {
				// 別スレッドの処理が終わるまで待つ。
				//<-q
			}

			// 長押しでリピートする機能。
			if false {
				// ここに来た場合信号出力が終わっている。
				// その時点で長押しされている場合複数回リピートしたい。
				// そのためここでリセットしておく。
				s.Now = PUSH_NONE
			}
		}
	}
}

func CheckChattering(s *PushState) bool {
	preState := s.Now
	s.Update()
	if s.Now == PUSH_NONE {
		return true
	}

	// 同時押しは無視する?
	// 長押しは無視する?

	// チャタリング防止のためデバウンス期間待つ。
	{
		f := s.Now != preState // ひとつ前の状態がfalseでこのループでtrueになったかを調べる。
		if !f {
			return true
		}
		//log.Printf("sleep for chattering: %vms", s.debounce)
		time.Sleep(s.debounce)
		// 再度チェック。
		s.Update()
		f = s.Now != preState && s.Now != PUSH_NONE
		if !f {
			return true
		}
	}
	return false // チャタリングしていなかった。
}

func ChangeSetting(s *PushState) bool {
	if s.Now == PUSH_ADD_SPEED {
		s.SpeedOffset++
		return true
	} else if s.Now == PUSH_SUB_SPEED {
		s.SpeedOffset--
		return true
	} else if s.Now == PUSH_ADD_FREQUENCY {
		s.FreqOffset++
		return true
	} else if s.Now == PUSH_SUB_FREQUENCY {
		s.FreqOffset--
		return true
	} else if s.Now == PUSH_ADD_DEBOUNCE {
		s.DebounceOffset++
		return true
	} else if s.Now == PUSH_SUB_DEBOUNCE {
		s.DebounceOffset--
		return true
	} else if s.Now == PUSH_REVERSE {
		s.Reverse = !s.Reverse
		return true
	}
	return false
}
