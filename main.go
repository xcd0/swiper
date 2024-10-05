package main

import (
	"fmt"
	"machine"
	"sync"
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

	_wait bool
	mutex sync.Mutex
)

func main() {
	var s PushState
	ch := make(chan ePushState, 1) // キー入力用channel 先行入力できるほど人間はつよくないので容量1。
	q := make(chan struct{})       // 処理完了フラグ用channel
	_wait = true

	go OutputSignal(&s, ch, q) // 出力信号作成スレッド    : channelから受け取り、信号を生成して出力する。
	LoopPinCheck(&s, ch, q)    // キー入力監視(メイン)スレッド: ピン状態を読み取り、channelに投げる。
}

func OutputSignal(s *PushState, ch chan ePushState, q chan struct{}) {
	// chに入ってくるまで何もしない。
	// 入ってきたらchに合わせて処理する。
	// 処理が終わったらqを通して終了を通知する。
	for {
		select {
		case ps := <-ch: // ePushStateが入ってくる。
			if ps == PUSH_DIT {
				if !s.Reverse {
					fmt.Printf(".")
					OutputSineWhileTick(s, 1)
				} else {
					fmt.Printf("-")
					OutputSineWhileTick(s, 3)
				}
			} else if ps == PUSH_DASH {
				if !s.Reverse {
					fmt.Printf("-")
					OutputSineWhileTick(s, 3)
				} else {
					fmt.Printf(".")
					OutputSineWhileTick(s, 1)
				}
			} else {
				//
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
	// log.Printf("OutputSine: start")
	// defer log.Printf("OutputSine: start")
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
		calced := make([]uint32, len(sinTable))
		for i := 0; i < len(sinTable); i++ {
			calced[i] = uint32(float32(pwm.Top()) * sinTable[i])
		}
		tick := 10
		steps := len(sinTable) / tick // 正弦波のステップ数
		stepDuration := time.Second / time.Duration(sineFrequency) / time.Duration(steps)

		for {
			select {
			case _ = <-q:
				// log.Printf("OutputSine: recieved q")
				// qに入って来たので終了する。
				return
			default:
				// 1周期分正弦波を出力する。
				for i := 0; i < len(sinTable); i += tick {
					pwm.Set(pwm_ch, calced[i])
					time.Sleep(stepDuration)
				}
			}
		}
	}
}

func OutputSineWhileTick(s *PushState, ticks int) {
	// log.Printf("OutputSineWhileTick: start")
	// defer log.Printf("OutputSineWhileTick: end")
	q := make(chan struct{})
	// log.Printf("OutputSineWhileTick: go OutputSine")
	go OutputSine(s.freq, q) // 正弦波を出力開始。

	// 指定tick数の期間待つ。
	// log.Printf("OutputSineWhileTick: wait %v ticks", ticks)
	end := time.Now().Add(time.Duration(ticks) * s.tick)
	for time.Now().Before(end) {
		time.Sleep(time.Millisecond * time.Duration(1)) // 1msごとにチェック
	}

	// log.Printf("OutputSineWhileTick: output end")
	q <- struct{}{}                       // 正弦波出力を終了する。
	time.Sleep(s.tick * time.Duration(1)) // 文字ごとの間隔 1tick空ける。

	// 単語の間は4tick空ける必要があるが、このプログラムでは単語の判断は無理なのでユーザーが頑張るものとする。
}

func LoopPinCheck(s *PushState, ch chan ePushState, q chan struct{}) {

	// 4tickの間何も押されていなければ空白1つだけを送出する。
	none := true
	end := time.Now().Add(time.Duration(4) * s.tick)

	for {
		// 1msごとにチェック。
		time.Sleep(time.Millisecond * time.Duration(1))
		if none && time.Now().After(end) {
			// 4tickの間何も押されていなければ空白1つだけを送出する。リピートはしない。
			fmt.Printf(" ")
			none = false
		}

		// チャタリングを防止しつつキー入力があるまで待機する。
		if CheckChattering(s) {
			continue // チャタリング防止の待機時間内にスイッチがOFFになった。
		}
		// ここに来た場合キーが入力されている。
		// ピンに従って処理する。

		// 設定値変更ピンの場合の処理。
		// 設定値変更ピンでない場合は信号を出力する。
		if ChangeSetting(s) {
			// 設定値変更ピンがONになっていた。設定値を変更したので終わる。
		} else if s.Now == PUSH_STRAIGHT {
			OutputStraightKey(s) // もしストレートキー用ピンが押されていたら押されている間出力する。つまりチャタリング防止だけしつつそのまま出力する。
		} else {
			// ストレートキー用キー以外(長音、単音)は適切に出力する必要がある。
			OutputSwipeKey(s, ch, q)
		}
		end = time.Now().Add(time.Duration(4) * s.tick)
		none = true
	}
}

func OutputSwipeKey(s *PushState, ch chan ePushState, q chan struct{}) {
	// log.Printf("OutputSwipeKey: start")
	//defer log.Printf("OutputSwipeKey: end")
	n := s.Now
	for {
		// 別スレッドで信号処理。
		// 具体的にはmainから呼ばれているOutputSignalで処理される。
		ch <- s.Now
		if _wait {
			// 別スレッドの処理が終わるまで待つ。
			// log.Printf("OutputSwipeKey: waiting")
			<-q
			// log.Printf("OutputSwipeKey: recieved q")
			//log.Printf("OutputSwipeKey: wait end")
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
	fmt.Printf("*")
	// 押されている間ONにする。
	q := make(chan struct{})
	go OutputSine(s.freq, q)
	for {
		// チェック
		s.Update()
		if s.Now != PUSH_STRAIGHT {
			break
		}
	}
	// OFFになったので終わる。
	q <- struct{}{}
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
