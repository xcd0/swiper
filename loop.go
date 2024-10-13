package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// ループで入力状態をチェックし、
// チャタリングを除去しながら、
// リングバッファrbに蓄積する。
func LoopPinCheck(ch chan struct{}, q chan struct{}) {
	// 4tickの間何も押されていなければ空白1つだけを送出する。
	//none := true
	//end := time.Now().Add(time.Duration(4) * s.tick)
	// 5秒間の間何も押されていなければ罫線と現状の設定を送出する。
	//end2 := time.Now().Add(time.Duration(5) * time.Second)
	//none2 := true

	// ターミナル表示用バッファ。何が押されたかをためておく。

	for {
		// 1msごとにチェック。
		time.Sleep(time.Millisecond * time.Duration(1))
		//PrintCharactor(&none, &end, buf)
		//PrintSetting(&none2, &end2)

		// チャタリングを防止しつつキー入力があるまで待機する。
		if CheckChattering() {
			continue // チャタリング防止の待機時間内にスイッチがOFFになった。
		}

		// ここに来た場合キーが入力されている。
		// ピンに従って処理する。

		// 設定値変更ピンの場合の処理。
		// 設定値変更ピンでない場合は信号を出力する。
		//if ChangeSetting() {
		//	// 設定値変更ピンがONになっていた。設定値を変更したので終わる。
		//	// この時点では設定値を保存しない。
		//	// なので設定値変更ピンで変更された直後電源がOFFになった場合変更は保存されない。
		//} else {
		//	if s.preSetting != s.setting {
		//		// 設定値が変更されていた。できるだけflashに書き込まないために、変更されるたびには保存していない。
		//		// 設定値が変更されている状態で、設定値変更ピン以外のピン(長短ピン、ストレートキーなど)が押されたタイミングで保存する。
		//		if err := writeSettingToFile(filesystem, s.setting, setting_filepath); err != nil { // 変更された設定をflashメモリに保存
		//			handleError(fmt.Errorf("main: %v", err))
		//		}
		//		s.preSetting = s.setting
		//	}

		//	 if s.Now == PUSH_STRAIGHT {
		//	 	OutputStraightKey(s) // もしストレートキー用ピンが押されていたら押されている間出力する。つまりチャタリング防止だけしつつそのまま出力する。
		//	 } else {
		//	 	// ストレートキー用キー以外(長音、単音)は適切に出力する必要がある。
		//	 	OutputSwipeKey(s, ch, q)
		//	 }
		//}
		//end = time.Now().Add(time.Duration(4) * s.tick)
		//end2 = time.Now().Add(time.Duration(5) * time.Second)
		//none = true
		//none2 = true
	}
}

func IsChattering() bool {
	// チャタリングの影響を除去する。
	s.Now = ReadGPIO()
	time.Sleep(time.Microsecond * time.Duration(s.setting.Debounce)) // チャタリング防止のためデバウンス期間待つ。
	prestate := s.Now
	if s.Now = ReadGPIO(); prestate == s.Now {
		if prestate == 0 && s.Now == 0 {
			// 何も押されていなかった。
		} else {
			return false // チャタリング防止期間の前後で入力に変化がなかった。受理する。
		}
	}
	// チャタリング防止時間の間に入力に変化があった。無視する。
	return true
}

func PinCheck() {
	check := func(t int) bool {
		return s.Now&1<<t != 0
	}
	var prestate PinState          // ひとつ前のループの状態。
	var straightKeyStart time.Time // ストレートキーが押下開始した時刻。
	var changeSetting bool = false // 設定変更があった
	for {
		time.Sleep(time.Millisecond * 1) // 1msごとにチェック。
		if IsChattering() {              // チャタリングの影響を除去する。
			continue
		}
		if prestate == 0 {
			if s.Now != 0 {
				// 何も押されていない状態で、何かが押された。
				if AddMorseKey(&straightKeyStart) {
					// モールス信号の入力の場合リングバッファに追加する。
					changeSetting = false
				} else {
					// それ以外の入力だった。
					if check(s.setting.PinSetting.Reset) { // 設定リセットピン。
						// 設定リセット。
						s.setting = NewSetting()
					} else if check(s.setting.PinSetting.Reverse) { // 長短反転ピン。
						s.setting.Reverse = !s.setting.Reverse
					}
				}
			} else {
				// 来ないはず。
			}
		} else {
			// 何かが押されている状態で、何も押されていない状態になった。
			// 何かが押されている状態で、それは離されて、他が押された。(ほぼあり得ない)
			// 何かが押されている状態で、何かが追加で押された。(同時押し)
			// 何かが押されている状態で、押された状態が維持された。(長押し)
			if s.Now == 0 {
				// 何かが押されている状態で、何も押されていない状態になった。
			} else if xor := prestate ^ s.Now; xor != 0 {
				// 何かが押されている状態で、何かが追加で押された。(同時押し)
			} else if xor == 0 {
				// 何かが押されている状態で、押された状態が維持された。(長押し)
				// 現状長押しで何か効果がある入力は、ストレートキーと設定変更のみ。
				if CheckStraightKey(&straightKeyStart) {
					// ストレートキーが押されていたのでストレートキーが離れるまでチェックして長音なのか短音なのか判定してリングバッファに追加。
				} else if CheckChangeSetting(prestate) {
					// 設定変更ピンが押されていたので設定値を更新した。
					changeSetting = true // 設定変更があった。次のモールス信号の打鍵までは設定変更をflashに書き込まない。
				}
			}
		}

		{
			// 1ループ終了時の処理。
			if !changeSetting && s.preSetting != s.setting {
				// 設定値が変更されていた。できるだけflashに書き込まないために、変更されるたびには保存していない。
				// 設定値が変更されている状態で、設定値変更ピン以外のピン(長短ピン、ストレートキーなど)が押されたタイミングで保存する。
				if err := writeSettingToFile(filesystem, s.setting, setting_filepath); err != nil { // 変更された設定をflashメモリに保存
					handleError(fmt.Errorf("main: %v", err))
				}
				s.preSetting = s.setting
			}
			prestate = s.Now
		}
	}
}

// if check(s.setting.PinSetting.DecodeCW) { // デコード用CW入力ピン。外部からのCWをでコードしてUSB Serialに文字として出力する。
//
//		(Push_DecodeCW)
//	}
//
// if check(s.setting.PinSetting.Reverse) { // 長短音ピン反転ピン。
//
//		(Push_Reverse)
//	} else if check(s.setting.PinSetting.Reset) { // 設定リセットピン。
//
//		(Push_Reset)
//	} else {
//
//		(Change_AnalogChangeSpeed)
//		(Change_AnalogChangeFrequency)
//	}
func CheckChangeSetting(prestate PinState) bool {

	// Fnピン以外はこの関数を呼ぶ前に処理している。
	// 以降はFnピンの未処理する。
	// Fn1-4が1つ押されているかつ、Fn5-6が1つ押されていることを前提とする。
	if !CheckPrerequisitesForSetting(prestate) {
		return false
	}

	// Fn1 (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録1を指定する。 / (1秒以上) 設定値変更:スピード変更状態にする。
	// Fn2 (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録2を指定する。 / (1秒以上) 設定値変更:正弦波周波数変更状態にする。
	// Fn3 (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録3を指定する。 / (1秒以上) 設定値変更:デバウンス変更状態にする。
	// Fn4 (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録4を指定する。 / (1秒以上) 設定値変更:長音比率変更状態にする。
	// Fn5 (1秒未満) 基本は何もしない。設定値変更状態で設定値を増加する。                / (1秒以上) 入力記録状態にする。長押ししながらFn1～4を短く押して記録先を指定する。その状態でFn5を離して打鍵すると記録される。再度Fn5を押して記録状態を終了する。1分以上記録終了されないとき記録を破棄して過去の記録状態を維持する。
	// Fn6 (1秒未満) 基本は何もしない。設定値変更状態で設定値を減少する。                / (1秒以上) 記録出力状態にする。設定ファイルに記述がある文字列を出力する。長押ししながらFn1～4を短く押して定型文を出力する。

	check := func(t int) bool {
		return s.Now&1<<t != 0
	}
	checkpre := func(t int, prestate PinState) bool {
		return prestate&1<<t != 0
	}
	pre_fn1 := checkpre(s.setting.PinSetting.Fn1, prestate)
	pre_fn2 := checkpre(s.setting.PinSetting.Fn2, prestate)
	pre_fn3 := checkpre(s.setting.PinSetting.Fn3, prestate)
	pre_fn4 := checkpre(s.setting.PinSetting.Fn4, prestate)
	pre_fn5 := checkpre(s.setting.PinSetting.Fn5, prestate)
	pre_fn6 := checkpre(s.setting.PinSetting.Fn6, prestate)
	now_fn1 := check(s.setting.PinSetting.Fn1)
	now_fn2 := check(s.setting.PinSetting.Fn2)
	now_fn3 := check(s.setting.PinSetting.Fn3)
	now_fn4 := check(s.setting.PinSetting.Fn4)
	now_fn5 := check(s.setting.PinSetting.Fn5)
	now_fn6 := check(s.setting.PinSetting.Fn6)

	if pre_fn1 { // 速度。初期値20WPM。
		if now_fn5 {
			s.setting.Speed++
		} else if now_fn6 {
			s.setting.Speed--
		}
	} else if pre_fn2 { // 正弦波周波数。初期値700Hz。
		if now_fn5 {
			s.setting.Frequency += 10
		} else if now_fn6 {
			s.setting.Frequency -= 10
		}
	} else if pre_fn3 { // デバウンス期間。初期値20s。
		if now_fn5 {
			s.setting.Debounce += 5
		} else if now_fn6 {
			s.setting.Debounce -= 5
		}
	} else if pre_fn4 { // 長音比率。初期値3.0。
		if now_fn5 {
			s.setting.DashRate += 0.1
		} else if now_fn6 {
			s.setting.DashRate -= 0.1
		}
	} else if pre_fn5 {
		// 入力記録状態にする。長押ししながらFn1～4を短く押して記録先を指定する。
		// その状態でFn5を離して打鍵すると記録される。再度Fn5を押して記録状態を終了する。1分以上記録終了されないとき記録を破棄して過去の記録状態を維持する。
		id := 0
		if now_fn1 {
			id = 0
		}
		if now_fn2 {
			id = 1
		}
		if now_fn3 {
			id = 2
		}
		if now_fn4 {
			id = 3
		}
		// 何も押されていな状態になるのを待つ。
		for {
			time.Sleep(time.Millisecond * time.Duration(1)) // 1msごとにチェック。
			s.Now = ReadGPIO()
			if AreAllPinsOff(s.Now) {
				break
			}
		}
		// 入力記録状態。
		forceBreak := true                 // 5分経っても入力記録状態が終わらない場合は破棄する。
		rrb := NewInputInputRingBuffer(20) // 1文字分のモールス信号を保持するリングバッファ。
		end := time.Now().Add(time.Second * 300)
		var last time.Time = time.Now() // 最後に受理した時刻
		var straightKeyStart time.Time  // ストレートキーが押下開始した時刻。
		var morseCode strings.Builder
		for time.Now().Before(end) {
			for {
				time.Sleep(time.Millisecond * time.Duration(1)) // 1msごとにチェック。
				if time.Now().Before(end) {
					break
				}
				if time.Now().Sub(last) > s.tick*4 {
					// 文字受理。
					if rrb.Count() != 0 {
						if decoded, err := DecodeMorse(rrb, s.setting.Mode); err != nil { //
							log.Printf("%v", err)
						} else {
							morseCode.WriteString(decoded)
						}
						morseCode.WriteString(" ")
					}
				}
				if IsChattering() { // チャタリングの影響を除去する。
					continue
				}
				if check(s.setting.PinSetting.Fn5) {
					forceBreak = false
					break
				} else {
					if check(s.setting.PinSetting.InputDit) { // 短音入力ピン。
						rrb.Add(Push_InputDit)
					} else if check(s.setting.PinSetting.InputDash) { // 長音入力ピン。
						rrb.Add(Push_InputDash)
					} else if CheckStraightKey(&straightKeyStart) {
					} else if s.Now == 0 {
						continue // 来ないはず。
					}
				}
				last = time.Now()
			}
		}
		if forceBreak {
			// 破棄する。
			return false
		} else {
			s.setting.Recorded[id] = morseCode.String()
		}
	} else if pre_fn6 {
		// 記録出力状態にする。設定ファイルに記述がある文字列を出力する。
		// 長押ししながらFn1～4を短く押して定型文を出力する。
		id := 0
		if now_fn1 {
			id = 0
		}
		if now_fn2 {
			id = 1
		}
		if now_fn3 {
			id = 2
		}
		if now_fn4 {
			id = 3
		}
		OutputString := func(str *string) {}
		OutputString(&s.setting.Recorded[id])
	}

	{
		// 変更された設定値が意図しない値にならないように修正しておく。
		s.setting.Speed = Clamp(s.setting.Speed, 1, 999)             // 速度。初期値20WPM。
		s.setting.Frequency = Clamp(s.setting.Frequency, 20, 200000) // 正弦波周波数。初期値700Hz。
		s.setting.Debounce = Clamp(s.setting.Debounce, 0, 10000)     // デバウンス期間。初期値20s。
		s.setting.DashRate = clamp(s.setting.DashRate, 3.0, 10.0)    // 長音比率。初期値3.0。
	}
	return true
}

func CheckStraightKey(straightKeyStart *time.Time) bool {
	check := func(t int) bool {
		return s.Now&1<<t != 0
	}
	if check(s.setting.PinSetting.InputAny) { // 任意入力ピン。ストレートキー用ピン。
		// いつまでストレートキーが押されるか調べる。
		// 離れるまでの間その他の入力は無視する。
		for {
			time.Sleep(time.Millisecond * time.Duration(1)) // 1msごとにチェック。
			s.Now = ReadGPIO()
			if !check(s.setting.PinSetting.InputAny) { // 任意入力ピン。ストレートキー用ピン。
				// ストレートキーが離された。
				d := time.Now().Sub(*straightKeyStart) // ストレートキーがONだった時間。
				// とりあえず、現在の設定のWPMで3tick以上長押しされていたらdash扱い、それ未満なら短点扱い。
				if d < s.tick*3 {
					rb.Add(Push_InputDit)
				} else { // 長音入力ピン。
					rb.Add(Push_InputDash)
				}
				break
			}
		}
		return true
	}
	return false
}

func AddMorseKey(straightKeyStart *time.Time) bool {
	check := func(t int) bool {
		return s.Now&1<<t != 0
	}
	if check(s.setting.PinSetting.InputDit) { // 短音入力ピン。
		rb.Add(Push_InputDit)
		return true
	} else if check(s.setting.PinSetting.InputDash) { // 長音入力ピン。
		rb.Add(Push_InputDash)
		return true
	} else if check(s.setting.PinSetting.InputAny) { // 任意入力ピン。ストレートキー用ピン。
		// rb.Add(Push_InputAny)
		// ここでは追加しない。
		*straightKeyStart = time.Now()
		return true
	}
	// リングバッファにはモールス信号の打鍵のみ入れる。
	return false
}

// func PrintCharactor(none *bool, end *time.Time, buf *[]ePushState) {
// 	if *none && time.Now().After(*end) {
// 		if true {
// 			// バッファにある長短を解析して文字に変換する。
// 			char := ReadBuf(buf)
// 			fmt.Printf("\t%v\r\n", char)
// 			*none = false
// 		} else {
// 			// 4tickの間何も押されていなければ空白1つだけを送出する。リピートはしない。
// 			fmt.Printf(char_space)
// 			*none = false
// 		}
// 	}
// }

func CheckChattering() bool {
	s.Now = ReadGPIO()
	if s.Now == 0 {
		return true
	}
	time.Sleep(time.Microsecond * time.Duration(s.setting.Debounce)) // チャタリング防止のためデバウンス期間待つ。
	if state := ReadGPIO(); s.Now == state {
		s.Now = state
		return true
	}
	return false
}

// func ChangeSetting(s *PushState) bool {
// 	s.preSetting = s.setting
// 	ret := false
// 	if s.Now == PUSH_RESET {
// 		// 設定初期化。
// 		s.setting.SpeedOffset = 0
// 		s.setting.FreqOffset = 0
// 		s.setting.DebounceOffset = 0
// 		s.setting.Reverse = false
// 		ret = true
// 	} else if s.Now == PUSH_ADD_SPEED {
// 		s.setting.SpeedOffset++
// 		ret = true
// 	} else if s.Now == PUSH_SUB_SPEED {
// 		s.setting.SpeedOffset--
// 		ret = true
// 	} else if s.Now == PUSH_ADD_FREQUENCY {
// 		s.setting.FreqOffset++
// 		ret = true
// 	} else if s.Now == PUSH_SUB_FREQUENCY {
// 		s.setting.FreqOffset--
// 		ret = true
// 	} else if s.Now == PUSH_ADD_DEBOUNCE {
// 		s.setting.DebounceOffset++
// 		ret = true
// 	} else if s.Now == PUSH_SUB_DEBOUNCE {
// 		s.setting.DebounceOffset--
// 		ret = true
// 	} else if s.Now == PUSH_REVERSE {
// 		s.setting.Reverse = !s.setting.Reverse
// 		ret = true
// 	}
//
// 	if ret {
// 		UpdateSetting(s)
// 	}
// 	return ret
//
// 	// 設定値変更
// 	// 設定値変更状態にするピンが押された後に設定値増減ピンが押される他ことを確認する。
// 	{
// 		only, macroState, pin := CheckFnPin()
// 		if !only {
// 			return
// 		}
//
// 		// 3段階ある。
// 		// 1. マクロピンが1つ押されている。
// 		// 2. マクロピンの状態が維持されている
// 		// 3. マクロピンが2つ押されている。
// 		// 3段階目になったら処理する。
// 		stateNum := 1
//
// 		for {
// 			time.Sleep(time.MilliSecond) // 1ms待つ。
// 			preFnState = macroState
// 			only, macroState, _ = CheckFnPin()
//
// 			if stateNum != 3 {
// 				if preFnState == macroState {
// 					if !(iniState & 1 << pin) {
// 						return // 長押ししていたマクロピンが離された。
// 					}
// 					continue // 1段階目
// 				}
// 				if only {
// 					if !(iniState & 1 << pin) {
// 						return // 長押ししていたマクロピンが離された。
// 					}
// 					stateNum = 2
// 					continue // 2段階目
// 				}
// 				// 3段階目になった。
// 				if stateNum == 2 {
// 					stateNum = 3
// 				}
// 			}
// 			if !(iniState & 1 << pin) {
// 				return // 長押ししていたマクロピンが離された。
// 			}
// 			if only {
// 				continue
// 			}
//
// 			// 長押ししていたピン以外にマクロピンが押されている。
// 			state := []bool{
// 				(s.Now & 1 << PinSetting.Fn1), // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録1を指定する。 / (同時押しで先に押す) 設定値変更:スピード変更状態にする。
// 				(s.Now & 1 << PinSetting.Fn2), // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録2を指定する。 / (同時押しで先に押す) 設定値変更:正弦波周波数変更状態にする。
// 				(s.Now & 1 << PinSetting.Fn3), // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録3を指定する。 / (同時押しで先に押す) 設定値変更:デバウンス変更状態にする。
// 				(s.Now & 1 << PinSetting.Fn4), // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録4を指定する。 / (同時押しで先に押す) 設定値変更:長音比率変更状態にする。
// 				(s.Now & 1 << PinSetting.Fn5), // (1秒未満) 基本は何もしない。設定値変更状態で設定値を増加する。                / (同時押しで先に押す) 入力記録状態にする。長押ししながらFn1～4を短く押して記録先を指定する。その状態でFn5を離して打鍵すると記録される。再度Fn5を押して記録状態を終了する。1分以上記録終了されないとき記録を破棄して過去の記録状態を維持する。
// 				(s.Now & 1 << PinSetting.Fn6), // (1秒未満) 基本は何もしない。設定値変更状態で設定値を減少する。                / (同時押しで先に押す) 記録出力状態にする。設定ファイルに記述がある文字列を出力する。長押ししながらFn1～4を短く押して定型文を出力する。
// 			}
// 			// state[pin] 以外のstateの何れかがtrueなはず。
//
// 			// マクロピン1-4とマクロピン5-6の1つづつが押されているはず。
// 			if (state[0] || state[1] || state[2] || state[3]) && (state[4] || state[5]) {
// 				continue // 無視する。
// 			}
//
// 			// 設定値を増減する。
// 			if pin < 4 {
// 				inc := s.Now & 1 << PinSetting.Fn5
// 				dec := s.Now & 1 << PinSetting.Fn6
// 				if pin == 0 { // 速度。初期値20WPM。
// 					if inc {
// 						s.setting.Speed++
// 					} else if dec {
// 						s.setting.Speed--
// 					}
// 					s.setting.Speed = Clamp(s.setting.Speed, 1, 999)
// 					// 1wpmは1分間にPARIS(50短点)を1回送る速さ。 例えば24wpmの短点は50ms、長点は150msになる。
// 					// つまり、n[wpm]は、1分間に(n*50)短点(1秒間にn*50/60短点)の速さなので、1短点は60/50/n*1000[ms]の長さになる。
// 					s.tick = time.Duration(1000*60/50/s.setting.Speed) * time.Millisecond
// 				} else if pin == 1 { // 正弦波周波数。初期値700Hz。
// 					if inc {
// 						s.setting.Frequency++
// 					} else if dec {
// 						s.setting.Frequency--
// 					}
// 					s.setting.Frequency = Clamp(s.setting.Frequency, 20, 100000)
// 				} else if pin == 2 { // デバウンス期間。初期値20us。
// 					if inc {
// 						s.setting.Debounce++
// 					} else if dec {
// 						s.setting.Debounce--
// 					}
// 					s.setting.Debounce = Clamp(s.setting.Debounce, 0, 10000)
// 				} else if pin == 3 { // 長音比率。初期値3.0。
// 					if inc {
// 						s.setting.DashRate++
// 					} else if dec {
// 						s.setting.DashRate--
// 					}
// 					s.setting.DashRate = clamp(s.setting.DashRate, 3, 10)
// 				}
// 				// ここに来たということは設定値が増減された。
// 				// 長押しされていたらリピートしたいので少し待つ。
// 				time.Sleep(time.Duration(200) * time.MilliSecond) // 200ms待つ。
// 			} else if pin < 6 {
// 				record_num := -1
// 				if state[0] {
// 					record_num = 0
// 				} else if state[1] {
// 					record_num = 1
// 				} else if state[2] {
// 					record_num = 2
// 				} else if state[3] {
// 					record_num = 3
// 				}
// 				if record_num == -1 {
// 					// 謎。
// 				} else {
// 					if pin == 4 {
// 						// 未実装。
// 						// 入力記録状態にする。長押ししながらFn1～4を短く押して記録先を指定する。
// 						// その状態でFn5を離して打鍵すると記録される。
// 						// 再度Fn5を押して記録状態を終了する。
// 						// 1分以上記録終了されないとき記録を破棄して過去の記録状態を維持する。
// 						end := time.Now().Add(time.Duration(60) * time.Second)
// 						for time.Now().Before(end) {
// 							// 長押しされていたピンが離されたか調べる。
// 							time.Sleep(time.MilliSecond) // 1ms待つ。
// 							only, macroState, p = CheckFnPin()
// 							if macroState & 1 << pin {
// 								// 離された。入力記録状態になる。
// 								RecordInput(end)
// 								break
// 							}
// 						}
// 					} else if pin == 5 {
// 						// 記録出力状態にする。設定ファイルに記述がある文字列を出力する。
// 						// 長押ししながらFn1～4を短く押して定型文を出力する。
// 						OutputRecorded(record_num)
// 					}
// 				}
// 				// ここに来たということは設定値が増減された。
// 				// 長押しされていたらリピートしたいので少し待つ。
// 				time.Sleep(time.Duration(200) * time.MilliSecond) // 200ms待つ。
// 			} else {
// 				// 謎。
// 			}
// 		}
// 	}
// }

func RecordInput(end time.Time) {
	// 記録する。
	// 未実装。
	for time.Now().Before(end) {
	}

}
func OutputRecorded(record_num int) {
	// 記録された定型文を出力する。
	log.Printf("未実装。 : %v", s.setting.Recorded[record_num])
	// 未実装。
}

// // 戻り値は一つだけ押されているか、押されている状態、1つだけ押されているときのマクロピン番号。
// func CheckFnPin() (only bool, state uint8, pin uint8) {
//
// 	m := func() []bool {
// 		s.Now = ReadGPIO()
// 		return []bool{
// 			(s.Now & 1 << s.setting.PinSetting.Fn1) != 0, // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録1を指定する。 / (同時押しで先に押す) 設定値変更:スピード変更状態にする。
// 			(s.Now & 1 << s.setting.PinSetting.Fn2) != 0, // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録2を指定する。 / (同時押しで先に押す) 設定値変更:正弦波周波数変更状態にする。
// 			(s.Now & 1 << s.setting.PinSetting.Fn3) != 0, // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録3を指定する。 / (同時押しで先に押す) 設定値変更:デバウンス変更状態にする。
// 			(s.Now & 1 << s.setting.PinSetting.Fn4) != 0, // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録4を指定する。 / (同時押しで先に押す) 設定値変更:長音比率変更状態にする。
// 			(s.Now & 1 << s.setting.PinSetting.Fn5) != 0, // (1秒未満) 基本は何もしない。設定値変更状態で設定値を増加する。                / (同時押しで先に押す) 入力記録状態にする。長押ししながらFn1～4を短く押して記録先を指定する。その状態でFn5を離して打鍵すると記録される。再度Fn5を押して記録状態を終了する。1分以上記録終了されないとき記録を破棄して過去の記録状態を維持する。
// 			(s.Now & 1 << s.setting.PinSetting.Fn6) != 0, // (1秒未満) 基本は何もしない。設定値変更状態で設定値を減少する。                / (同時押しで先に押す) 記録出力状態にする。設定ファイルに記述がある文字列を出力する。長押ししながらFn1～4を短く押して定型文を出力する。
// 		}
// 	}()
//
// 	// マクロピンが1つだけ押されていることを確認する。
// 	f := func() bool {
// 		c := 0
// 		for i, b := range m {
// 			if b {
// 				pin = uint8(i)
// 				c++
// 			}
// 		}
// 		return c == 1
// 	}()
// 	if !f {
// 		// 非常にまれだが、1ms程度の誤差で同時にマクロピンが押されている。
// 		// どちらが先に押されたか判定が難しいので無視する。
// 		return
// 	}
//
// 	var macroState uint8 = 0 |
// 		1<<int(m[0] && !(m[2] && m[3] && m[4] && m[5] && m[6])) |
// 		1<<int(m[1] && !(m[1] && m[3] && m[4] && m[5] && m[6])) |
// 		1<<int(m[2] && !(m[1] && m[2] && m[4] && m[5] && m[6])) |
// 		1<<int(m[3] && !(m[1] && m[2] && m[3] && m[5] && m[6])) |
// 		1<<int(m[4] && !(m[1] && m[2] && m[3] && m[4] && m[6])) |
// 		1<<int(m[5] && !(m[1] && m[2] && m[3] && m[4] && m[5]))
//
// 	state = 0 |
// 		1<<int(m[0]) |
// 		1<<int(m[1]) |
// 		1<<int(m[2]) |
// 		1<<int(m[3]) |
// 		1<<int(m[4]) |
// 		1<<int(m[5])
//
// 	// 一つだけ押されているかどうか、押されている状態。
// 	only = macroState != 0
// 	return only, state, pin
// }

func CheckPrerequisitesForSetting(prestate PinState) bool {
	// Fn1-4が1つ押されているかつ、Fn5-6が1つ押されていることを前提とする。
	pinFn14 := func(state PinState) int {
		ret := 0
		if state&1<<s.setting.PinSetting.Fn1 != 0 {
			ret++
		}
		if state&1<<s.setting.PinSetting.Fn2 != 0 {
			ret++
		}
		if state&1<<s.setting.PinSetting.Fn3 != 0 {
			ret++
		}
		if state&1<<s.setting.PinSetting.Fn4 != 0 {
			ret++
		}
		return ret
	}
	pinFn56 := func(state PinState) int {
		ret := 0
		if state&1<<s.setting.PinSetting.Fn5 != 0 {
			ret++
		}
		if state&1<<s.setting.PinSetting.Fn6 != 0 {
			ret++
		}
		return ret
	}
	{
		count_f14 := pinFn14(prestate)
		count_f56 := pinFn56(s.Now)
		if count_f14 == 1 && count_f56 == 1 {
			return true
		}
	}
	{
		count_f14 := pinFn14(s.Now)
		count_f56 := pinFn56(prestate)
		if count_f14 == 1 && count_f56 == 1 {
			return true
		}
	}
	return false
}
