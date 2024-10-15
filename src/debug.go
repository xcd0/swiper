package main

import (
	"device/rp"
	"fmt"
	"log"
	"machine"
	"time"
)

func debug() {
	// https://jhalfmoon.com/dbc/2023/10/23/go%E3%81%AB%E3%81%84%E3%82%8C%E3%81%B0go%E3%81%AB%E5%BE%93%E3%81%8839-%E3%83%A9%E3%82%BA%E3%83%91%E3%82%A4pico%E3%81%A7%E3%82%82%E3%83%AC%E3%82%B8%E3%82%B9%E3%82%BF%E7%9B%B4%E6%8E%A5%E3%82%A2/
	cpuidIMP := rp.PPB.GetCPUID_IMPLEMENTER()
	cpuidVAR := rp.PPB.GetCPUID_VARIANT()
	cpuidARCH := rp.PPB.GetCPUID_ARCHITECTURE()
	cpuidPNO := rp.PPB.GetCPUID_PARTNO()
	cpuidREV := rp.PPB.GetCPUID_REVISION()
	chipREV := rp.SYSINFO.GetCHIP_ID_REVISION()
	chipPART := rp.SYSINFO.GetCHIP_ID_PART()
	chipMANU := rp.SYSINFO.GetCHIP_ID_MANUFACTURER()

	end := time.Now().Add(time.Second * 30)

	for {
		fmt.Printf(`
# CPUID
# 	IMPLEMENTER:  %08x
# 	VARIANT:      %08x
# 	ARCHITECTURE: %08x
# 	PARTNO:       %08x
# 	REVISION:     %08x
# CHIP_ID
# 	REVISION:     %08x
# 	PART:         %08x
# 	MANUFACTURER: %08x
`, cpuidIMP, cpuidVAR, cpuidARCH, cpuidPNO, cpuidREV, chipREV, chipPART, chipMANU)
		time.Sleep(time.Second * 4)

		if time.Now().Before(end) {
			machine.EnterBootloader()
		}
	}
}

func debug_func() {

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
	_ = func(f int) {
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
		pin_beep_out := gpio[26]
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
	}
}
