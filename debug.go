package main

import (
	"log"
	"machine"
	"time"
)

func debug_func() {

	if false {
		_wait = true // 信号出力スレッドの終了を待機する

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
	if true {
		go func(sineFrequency int) {
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
				for i := 0; i < len(sinTable); i += tick {
					pwm.Set(pwm_ch, calced[i])
					time.Sleep(stepDuration)
				}
				continue
			}
		}(1000)
	} else if true {
		go func(f int) {
			pwm := machine.PWM0 // GPIO1でPWMする場合PWM0を指定すればよい。
			pwm.Configure(machine.PWMConfig{
				// 300Hzくらいで、1e6はだめだった。1e5は直列に抵抗を入れれば大丈夫。
				//Period: uint64(1e4), // 1e4は周波数が5kHzくらいになるとだめだった。
				//Period: uint64(1e3), // 400HzでFFTしたところこれが一番きれいな正弦波だった。
				Period: uint64(5e2),
				//Period: uint64(1e2),
				// Period: uint64(2e2), // 50kHzくらい
				//Period: uint64(1e2), // 100kHzくらい
				//Period: uint64(1e2), // 周波数が5kHzで1e2はだいぶ波形が雑になった。
			})
			var err error
			pwm_ch, err = pwm.Channel(pin_beep_out)
			if err != nil {
				println(err.Error())
				return
			}
			calced := make([]uint32, len(sinTable))
			for i := 0; i < len(sinTable); i++ {
				calced[i] = uint32(float32(pwm.Top()) * sinTable[i])
			}

			// サンプリングを減らす。
			//tick := 4
			//tick := 20
			tick := 10
			// サンプリング数によってループの最低時間(遅延を入れないときの1ループの時間)が変わる。
			// 遅延なしで、tickが20の時100kHz (1周期10us)
			// 遅延なしで、tickが10の時50kHz (1周期20us)
			// 遅延なしで、tickが5の時26kHz (1周期38461ns)
			// 遅延なしで、tickが4の時21kHz (1周期47619ns)
			// 遅延なしで、tickが2の時10.5kHz (1周期47619ns)
			// の周波数になった。 (可聴域を20kHzまでとすれば、最低でも5にする必要があるがまあそんなに高い周波数いらない。)
			// したがってtickの値によって出せる周波数に上限がある。

			sineFrequency := f
			steps := len(sinTable) / tick // 正弦波のステップ数

			period := time.Second / time.Duration(sineFrequency)
			// 1000Hzの時、708Hzになった。(1周期1412us)
			stepDuration := period / time.Duration(steps)
			log.Printf("stepDuration:%v", stepDuration)

			//stepDuration := period / (time.Duration(steps))
			//stepDuration := period / (time.Duration(steps)) - time.Duration(47619))
			//stepDuration := period / (time.Duration(steps)) //- time.Duration(47619))
			//log.Printf("stepDuration:%v, steps:%v, period:%v", stepDuration, steps, period)

			for {
				period = time.Second / time.Duration(sineFrequency)
				stepDuration = period / time.Duration(steps)
				log.Printf("stepDuration:%v, steps:%v, period:%v\n", stepDuration, steps, period)
				end := time.Now().Add(time.Duration(3) * time.Second)
				for time.Now().Before(end) {
					for i := 0; i < len(sinTable); i += tick {
						pwm.Set(pwm_ch, calced[i])
						time.Sleep(stepDuration)
					}
				}
				continue
				period = time.Second/time.Duration(sineFrequency) - time.Duration(200)*time.Microsecond
				stepDuration = period / time.Duration(steps)
				log.Printf("stepDuration:%v, steps:%v, period:%v\n", stepDuration, steps, period)
				end = time.Now().Add(time.Duration(3) * time.Second)
				for time.Now().Before(end) {
					for i := 0; i < len(sinTable); i += tick {
						pwm.Set(pwm_ch, calced[i])
						time.Sleep(stepDuration)
					}
				}
			}
		}(1000)
	} else if true {
		go func() {
			// ここはモニター用ビープ音の出力の設定。
			// PWM変調で正弦波を出力する。
			// 可聴域以上の周波数でPWM変調すれば行けるはず。
			// ダメなら回路にちょっとLPF入れる。

			// 搬送波周波数の設定。
			// 搬送波周波数をとりあえず10kHzとする。
			pwmFrequency := 10 * 1e+3 // Hz

			// PWMの設定。
			// ここはビープ音を出力するピンごとに代わってくる。
			pwm := machine.PWM0 // GPIO1でPWMする場合PWM0を指定すればよい。
			log.Printf("pwm: %T : %#q", pwm)

			pwm.Configure(machine.PWMConfig{
				Period: uint64(1e9 / pwmFrequency), // nano secondで指定する。 // https://github.com/tinygo-org/tinygo/blob/release/src/machine/pwm.go
			})
			var err error
			pwm_ch, err = pwm.Channel(pin_beep_out)
			if err != nil {
				println(err.Error())
				return
			}

			// 出力する周波数の正弦波を、搬送波周波数でサンプリング。
			// 毎度sinを生成するのは馬鹿らしいのでスライスにする。go generateで var sinTable []uint16が1000個生成される。
			// この個数は出力する正弦波の滑らかさに影響する。
			// PWMの搬送波周波数と正弦波の周波数の 比率でPWMパルスの数の目安が決まる。
			// 周波数がある程度ユーザーによって変更されることを考えて、1000分割する。
			// 正弦波の値を毎度計算するのはマイコン的にやばいので、lookup tableを用意しておく。
			// これは go genarate で生成する。

			steps := len(sinTable) // 正弦波のステップ数
			sineFrequency := 1000  // 仮
			period := time.Second / time.Duration(sineFrequency)
			stepDuration := period / time.Duration(steps)
			for {
				for i := 1; i < steps; i++ {
					pwm.Set(pwm_ch, uint32(sinTable[i])) // https://github.com/tinygo-org/tinygo/blob/ef4f46f1d1550beb62324d750c496b2b4a7f76d0/src/machine/machine_rp2040_pwm.go#L175
					time.Sleep(stepDuration)
				}
			}
		}()
	}
}
