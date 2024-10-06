package main

import (
	"fmt"
	"machine"
	"time"
)

func OutputSignal(s *PushState, ch chan ePushState, q chan struct{}, buf *[]ePushState) {
	// chに入ってくるまで何もしない。
	// 入ってきたらchに合わせて処理する。
	// 処理が終わったらqを通して終了を通知する。
	for {
		select {
		case ps := <-ch: // ePushStateが入ってくる。
			//log.Printf("OutputSignal: %v", ps)
			if ps == PUSH_DIT {
				// 短音
				if !s.setting.Reverse {
					fmt.Printf(char_dit)
					OutputSineWhileTick(s, 1)
					*buf = append(*buf, PUSH_DIT)
				} else {
					// 反転設定がONだったので長音を出力。
					fmt.Printf(char_dash)
					*buf = append(*buf, PUSH_DASH)
					OutputSineWhileTick(s, 3)
				}
			} else if ps == PUSH_DASH {
				// 長音
				if !s.setting.Reverse {
					fmt.Printf(char_dash)
					OutputSineWhileTick(s, 3)
					*buf = append(*buf, PUSH_DASH)
				} else {
					// 反転設定がONだったので短音を出力。
					fmt.Printf(char_dit)
					*buf = append(*buf, PUSH_DIT)
					OutputSineWhileTick(s, 1)
				}
			} else {
				// 設定値変更ピンなど。
				fmt.Printf("%v", ps)
				*buf = append(*buf, ps)
			}
			if _wait {
				// メインスレッドの待機を終了する。
				q <- struct{}{}
			}
		default:
			time.Sleep(time.Millisecond) // 1ms間隔でチェックする。
		}
	}
}

// goroutineとして呼び出すこと。
func OutputSine(sineFrequency int, q chan struct{}) {
	if sineFrequency < 1 {
		// 引数がおかしい
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	pin_out.High()
	led.High()
	defer pin_out.Low()
	defer led.Low()
	// qに入ってくるまで正弦波を出力する。
	{
		pwm := machine.PWM0 // GPIO1でPWMする場合PWM0を指定すればよい。
		pwm.Configure(machine.PWMConfig{Period: uint64(5e2)})
		var err error
		if pwm_ch, err = pwm.Channel(pin_beep_out); err != nil {
			println(err.Error())
			return
		}
		pick := 10
		steps := len(sinTable) / pick // 正弦波のステップ数

		stepDuration := time.Second / time.Duration(sineFrequency) / time.Duration(steps)
		for {
			select {
			case _ = <-q:
				// qに入って来たので終了する。
				return
			default:
				// 1周期分正弦波を出力する。
				for i := 0; i < len(sinTable); i += pick {
					pwm.Set(pwm_ch, calced[i])
					time.Sleep(stepDuration)
				}
			}
		}
	}
}

func OutputSineWhileTick(s *PushState, ticks int) {
	q := make(chan struct{})
	go OutputSine(s.freq, q) // 正弦波を出力開始。

	// 指定tick数の期間待つ。
	end := time.Now().Add(time.Duration(ticks) * s.tick)
	for time.Now().Before(end) {
		time.Sleep(time.Millisecond * time.Duration(1)) // 1msごとにチェック
	}

	q <- struct{}{}                       // 正弦波出力を終了する。
	time.Sleep(s.tick * time.Duration(1)) // 文字ごとの間隔 1tick空ける。

	// 単語の間は4tick空ける必要があるが、このプログラムでは単語の判断は無理なのでユーザーが頑張るものとする。
}

func OutputSwipeKey(s *PushState, ch chan ePushState, q chan struct{}) {
	n := s.Now
	for {
		// 別スレッドで信号処理。
		// 具体的にはmainから呼ばれているOutputSignalで処理される。
		ch <- s.Now
		if _wait {
			// 別スレッドの処理が終わるまで待つ。
			<-q
		}
		// 長押しでリピートする。
		// 出力が終わった時点で再度チェックしてまだ押されていたら再度出力する。
		// ピン状態が変わっていたら終了する。
		if s.Update(); n != s.Now {
			break
		}
	}
}

func OutputStraightKey(s *PushState) {
	// もしストレートキー用ピンが押されていたら押されている間出力する。つまりチャタリング防止だけしつつそのまま出力する。
	preState := s.Now
	fmt.Printf(char_straight)
	// 押されている間ONにする。
	q := make(chan struct{})
	go OutputSine(s.freq, q)
	for {
		// チェック
		s.Update()
		if s.Now != preState {
			break
		}
		time.Sleep(time.Millisecond)
	}
	// OFFになったので終わる。
	q <- struct{}{}
}
