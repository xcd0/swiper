package main

import (
	"fmt"
	"log"
	"machine"
	"strings"
	"sync"
	"time"
)

const (
	char_dit      = "." // 短音
	char_dash     = "-" // 長音
	char_space    = " " // 文字区切り
	char_straight = "*" // 任意長さストレートキー
)

var (
	gpio              []machine.Pin
	led               machine.Pin // 基板上のLED
	pin_out           machine.Pin // 出力ピン
	pin_beep_out      machine.Pin // モニター用サウンド出力ピン
	pin_dit           machine.Pin // 短音ピン
	pin_dash          machine.Pin // 長音ピン
	pin_straight      machine.Pin // ストレートキー用ピン
	pin_reset         machine.Pin // 設定リセットピン
	pin_add_speed     machine.Pin // スピードアップピン
	pin_sub_speed     machine.Pin // スピードダウンピン
	pin_add_frequency machine.Pin // 周波数アップピン
	pin_sub_frequency machine.Pin // 周波数ダウンピン
	pin_add_debounce  machine.Pin // デバウンスアップピン
	pin_sub_debounce  machine.Pin // デバウンスダウンピン
	pin_reverse       machine.Pin // パドルの長短切り替えピン

	pwm_ch     uint8
	calced     []uint32 // 正弦波のルックアップテーブルから計算したもの
	preSetting Setting  // 直前の設定状態。設定をflashに書き込むかどうかの判定に使う。できるだけflashに書き込みたくない。

	_wait bool
	mutex sync.Mutex
)

func main() {
	var s PushState
	{
		//log.Printf("main: load setting")
		// 設定を読み込み
		savedSetting, isValid, err := readSettingFromFile(filesystem, setting_filepath)
		if err != nil {
			log.Printf("main: %v\n", err)
		}
		if !isValid {
			fmt.Printf("No valid data found. Using default settings.\r\n")
		}
		s.setting = savedSetting
		preSetting = s.setting
		UpdateSetting(&s) // 読み込んだ設定値からその他の設定値を計算する。
		//log.Printf("main: loaded setting")
	}

	ch := make(chan ePushState, 1) // キー入力用channel 先行入力できるほど人間はつよくないので容量1。
	q := make(chan struct{})       // 処理完了フラグ用channel
	buf := make([]ePushState, 0, 20)
	_wait = true

	go OutputSignal(&s, ch, q, &buf) // 出力信号作成スレッド    : channelから受け取り、信号を生成して出力する。
	LoopPinCheck(&s, ch, q, &buf)    // キー入力監視(メイン)スレッド: ピン状態を読み取り、channelに投げる。
}

func LoopPinCheck(s *PushState, ch chan ePushState, q chan struct{}, buf *[]ePushState) {

	// 4tickの間何も押されていなければ空白1つだけを送出する。
	none := true
	end := time.Now().Add(time.Duration(4) * s.tick)
	// 5秒間の間何も押されていなければ罫線と現状の設定を送出する。
	end2 := time.Now().Add(time.Duration(5) * time.Second)
	none2 := true

	// ターミナル表示用バッファ。何が押されたかをためておく。

	for {
		// 1msごとにチェック。
		time.Sleep(time.Millisecond * time.Duration(1))
		if none && time.Now().After(end) {
			if true {
				// バッファにある長短を解析して文字に変換する。
				char := ReadBuf(buf)
				fmt.Printf("\t%v\r\n", char)
				none = false
			} else {
				// 4tickの間何も押されていなければ空白1つだけを送出する。リピートはしない。
				fmt.Printf(char_space)
				none = false
			}
		}
		if none2 && time.Now().After(end2) {
			str := fmt.Sprintf(
				`
----------------------------------------
Reverse        : %v	: 長音キーと短音キーを反転させるか
SpeedOffset    : %v	: (so) 速度設定値       変更可能 -20から+10まで
FreqOffset     : %v	: (fo) 周波数設定値     変更可能 -8から+8まで
DebounceOffset : %v	: (do) デバウンス設定値 変更可能 -2から+18まで
WPM            : %v	: WPM                      25+so [wpm]
Tick           : %v	: 短音の長さ               60/50/wpm*1000 [ms]
Freq           : %vHz	: 正弦波の周波数の目安 800+50*fo [Hz]
Debounce       : %v	: チャタリング防止時間     20+10*do [ns]
----------------------------------------
`,
				s.setting.Reverse,
				s.setting.SpeedOffset,
				s.setting.FreqOffset,
				s.setting.DebounceOffset,
				s.setting.SpeedOffset+25,
				s.tick,
				s.freq,
				s.debounce,
			)
			str = strings.ReplaceAll(str, "\n", "\r\n")
			fmt.Printf("%v", str)
			none2 = false
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
			// この時点では設定値を保存しない。
			// なので設定値変更ピンで変更された直後電源がOFFになった場合変更は保存されない。
		} else {

			if preSetting != s.setting {
				// 設定値が変更されていた。できるだけflashに書き込まないために、変更されるたびには保存していない。
				// 設定値が変更されている状態で、設定値変更ピン以外のピン(長短ピン、ストレートキーなど)が押されたタイミングで保存する。
				if err := writeSettingToFile(filesystem, s.setting, setting_filepath); err != nil { // 変更された設定をflashメモリに保存
					handleError(fmt.Errorf("main: %v", err))
				}
				preSetting = s.setting
			}

			if s.Now == PUSH_STRAIGHT {
				OutputStraightKey(s) // もしストレートキー用ピンが押されていたら押されている間出力する。つまりチャタリング防止だけしつつそのまま出力する。
			} else {
				// ストレートキー用キー以外(長音、単音)は適切に出力する必要がある。
				OutputSwipeKey(s, ch, q)
			}
		}
		end = time.Now().Add(time.Duration(4) * s.tick)
		end2 = time.Now().Add(time.Duration(5) * time.Second)
		none = true
		none2 = true
	}
}

func CheckChattering(s *PushState) bool {
	s.Update()
	if s.Now == PUSH_NONE {
		return true
	}

	// 同時押しは無視する?
	// 長押しは無視する?

	// チャタリング防止のためデバウンス期間待つ。
	{
		if s.Now == PUSH_NONE {
			return true
		}
		// デバウンスのために設定期間の間待つ。
		time.Sleep(s.debounce)
		s.Update() // 再度チェック。
		if s.Now == PUSH_NONE {
			return true
		}
	}
	// チャタリングしていなかった。
	return false
}

func ChangeSetting(s *PushState) bool {
	preSetting = s.setting
	ret := false
	if s.Now == PUSH_RESET {
		// 設定初期化。
		s.setting.SpeedOffset = 0
		s.setting.FreqOffset = 0
		s.setting.DebounceOffset = 0
		s.setting.Reverse = false
		ret = true
	} else if s.Now == PUSH_ADD_SPEED {
		s.setting.SpeedOffset++
		ret = true
	} else if s.Now == PUSH_SUB_SPEED {
		s.setting.SpeedOffset--
		ret = true
	} else if s.Now == PUSH_ADD_FREQUENCY {
		s.setting.FreqOffset++
		ret = true
	} else if s.Now == PUSH_SUB_FREQUENCY {
		s.setting.FreqOffset--
		ret = true
	} else if s.Now == PUSH_ADD_DEBOUNCE {
		s.setting.DebounceOffset++
		ret = true
	} else if s.Now == PUSH_SUB_DEBOUNCE {
		s.setting.DebounceOffset--
		ret = true
	} else if s.Now == PUSH_REVERSE {
		s.setting.Reverse = !s.setting.Reverse
		ret = true
	}

	if ret {
		UpdateSetting(s)
	}
	return ret
}

func UpdateSetting(s *PushState) {
	// 設定値からその他の設定値を計算する。
	{ // スピード
		s.setting.SpeedOffset = Clamp(s.setting.SpeedOffset, -20, 10) // とりあえず5wpmから35wpmまでとする。
		wpm := (s.setting.SpeedOffset + 25)                           // とりあえず初期値を25wpmとする。wpmから1つの短音の長さを計算する。
		// 1wpmは1分間にPARIS(50短点)を1回送る速さ。 例えば24wpmの短点は50ms、長点は150msになる。
		// つまり、n[wpm]は、1分間に(n*50)短点(1秒間にn*50/60短点)の速さなので、1短点は60/50/n*1000[ms]の長さになる。
		s.tick = time.Duration(1000*60/50/wpm) * time.Millisecond
	}
	{ // 音程
		s.setting.FreqOffset = Clamp(s.setting.FreqOffset, -8, 8) // とりあえず、800を基準に、400から1200までを狙って50Hz刻みとする。
		s.freq = (s.setting.FreqOffset*50 + 800)                  // とりあえず、初期値を800Hzとする。
	}
	{ // デバウンス
		s.setting.DebounceOffset = Clamp(s.setting.DebounceOffset, -2, 18) // 20msを基準に0msから200msまでを狙って10ms刻みとする。
		s.debounce = time.Duration(s.setting.DebounceOffset*10 + 20)
	}
}
