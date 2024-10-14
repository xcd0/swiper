package main

import (
	"fmt"
	"log"
	"machine"
	"time"
)

const (
	char_dit      = "." // 短音
	char_dash     = "-" // 長音
	char_space    = " " // 文字区切り
	char_straight = "*" // 任意長さストレートキー

	long_press = time.Duration(200) * time.Millisecond // 長押し扱いする秒数

	_debug = true
)

// machineのPWMは非公開なので。
type PWM interface {
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
	Enable(bool)
	Top() uint32
	Configure(config machine.PWMConfig) error
	Channel(machine.Pin) (uint8, error)
}

var (
	//s    PushState
	gpio []machine.Pin
	//led             machine.Pin                                    // 基板上のLED。
	//rb              *InputRingBuffer = NewInputInputRingBuffer(20) // 1文字分のモールス信号を保持するリングバッファ。
	//mutex_rb        sync.Mutex
	//mutex_sine      sync.Mutex
	pwm_ch          uint8
	pwm_for_monitor PWM      // モニター用正弦波出力用PWM。
	calced          []uint32 // 正弦波のルックアップテーブルから計算したもの。
	//port_map        map[int]string // どの番号のGPIOがどの用途で使用されているかのmap。デバッグ用。
	//input_scan_rate time.Duration  // 入力ピンの変化を読み取る周期。

	fc uint64 = 100 * 1e3 // 搬送波の周波数
	f  int    = 700       // 出力する正弦波の周波数

)

func main() {

	/*
		pwmはカウンターで成り立っている。
		周波数を上げると1波長分のカウンターが小さくなる。
		例えば125MHzのCPUで15.625MHz(125MHz/8)のPWMを行うと、PWMの分解能が8になる。
		正弦波を出力するときの量子化精度を128(7bit)にするの場合、976.5625(125MHz/128)kHzになる。
		つまり、正弦波の滑らかさと、PWM変調の搬送波の周波数はトレードオフの関係にある。
		出力したい正弦波の周波数fを20kHzまで再生できるようにする場合、
		搬送波周波数fcをfで割ったfc/fが、出力したい正弦波1波長分をPWM出力何回で行うかに対応する。
		20kHzの1波長を100分割してPWM出力するには最低2MHz(20kHz*100)のfcが必要になる。
		10分割なら最低200kHzあればよい。
		逆にfcを固定してfを変化させる。fcが200kの時、1周期tc(1/fc)は5us(5e-6)なのでこれを固定する。
		fを1kHzとすると、1周期t(1/f)をtcで分割して200(1e-3/5e-6)分割して出力することになる。
	*/
	{
		{
			pwm_pin := gpio[26]
			pwm_pin.Configure(machine.PinConfig{Mode: machine.PinPWM})
			pwms := [...]PWM{machine.PWM0, machine.PWM1, machine.PWM2, machine.PWM3, machine.PWM4, machine.PWM5, machine.PWM6, machine.PWM7}
			slice, err := machine.PWMPeripheral(pwm_pin)
			if err != nil {
				handleError(fmt.Errorf("init_pwm: %v", err))
				return
			}
			pwm_for_monitor = pwms[slice]
			if err = pwm_for_monitor.Configure(machine.PWMConfig{Period: 1e9 / fc}); err != nil {
				handleError(fmt.Errorf("init_pwm: %v", err))
				return
			}
			if pwm_ch, err = pwm_for_monitor.Channel(pwm_pin); err != nil {
				handleError(fmt.Errorf("init_pwm: %v", err))
				return
			}
			calced = make([]uint32, len(sinTable))
			for i := 0; i < len(sinTable); i++ {
				calced[i] = uint32(float32(pwm_for_monitor.Top()) * sinTable[i])
			}
		}

		debug()
	}
}

func debug() {

	// 出力する正弦波の1波長を何分割して出力するかを計算。
	pick := int((1. / float64(f)) / (1. / float64(fc)))                    // 分割数
	step := len(sinTable) / pick                                           // 正弦波を1000分割したlookuptableを用意しているので、これをpick数で分割した1stepを計算する。
	stepDuration := time.Second / (time.Duration(f) * time.Duration(pick)) // 1ステップの期間。

	for {
		t := time.Now()
		for j := 0; j < 1000; j++ {
			for i := 0; i < len(sinTable); i += step {
				pwm_for_monitor.Set(pwm_ch, calced[i])
				time.Sleep(stepDuration)
			}
		}
		log.Printf("t:%v", time.Now().Sub(t))
	}
}

// エラー時の無限ループ処理
func handleError(err error) {
	for {
		fmt.Printf("%v\r\n", err)
		time.Sleep(time.Second)
	}
}
