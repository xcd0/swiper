package main

import (
	"fmt"
	"strings"
	"time"
)

func UpdateSetting() {
	s.Now := ReadGPIO()
}

// 外部から変更可能な設定値。変更時にflashに保存する。起動時に読み込みを試みる。
type Setting struct {
	PinSetting PinAssign `json:"pin-setting"` // GPIOのピン設定。

	Reverse   bool    `json:"paddle-reverse"` // 長短のパドル反転。初期値false。
	Speed     int     `json:"speed"`          // 速度。初期値20WPM。
	Frequency int     `json:"frequency"`      // 正弦波周波数。初期値700Hz。
	Debounce  int     `json:"debounce"`       // デバウンス期間。初期値20s。
	DashRate  float32 `json:"dash-rate"`      // 長音比率。初期値3.0。

	// 負論理を使用するか。初期値false。
	EnableNegativeLogic bool `json:"enable-negative-logic"`

	// アナログ出力による制限は出力を試みるか。初期値false。
	// falseの時、モニター用制限は出力においてPWM変調によって出力する。
	// trueの時、モニター用出力ピンは GPIO26, GPIO27, GPIO28, GPIO29 の何れかに設定する必要がある。そうでないときPWM出力する。
	EnableAnalogOutput bool `json:"enable-analog-output-for-monitor"`

	// 長短パドルを同時に押したときのスクイーズ機能を無効化するか。初期値false。
	DisableSqueeze bool `json:"disable-squeeze"`
	// 自動で無入力期間を入れる機能を無効化するか。初期値false。
	DisableAutoSpace bool `json:"disable-auto-space"`
}

func PrintSetting(none *bool, end *time.Time) {
	if *none && time.Now().After(*end) {
		str := fmt.Sprintf(
			`
----------------------------------------------
Rev: %v	| in dit : %2d | out    : %2d | fn1 : %2d
WPM: %v	| in dash: %2d | out sin: %2d | fn2 : %2d
Tic: %v	| in any : %2d | ch wpm : %2d | fn3 : %2d
Frq: %v	| decode : %2d | ch freq: %2d | fn4 : %2d
Deb: %v	| reverse: %2d |              | fn5 : %2d
DaR: %v	|              | I2C SDA: %2d | fn6 : %2d
ASp: %v	| reset  : %2d | I2C SCL: %2d |          
----------------------------------------------
`,
			s.setting.Reverse, s.PinSetting.InputDit, s.PinSetting.Output, s.PinSetting.Fn1,
			s.setting.Speed, s.PinSetting.InputDash, s.PinSetting.OutputSine, s.PinSetting.Fn2,
			s.tick, s.PinSetting.InputAny, s.PinSetting.AnalogChangeSpeed, s.PinSetting.Fn3,
			s.setting.Frequency, s.PinSetting.DecodeCW, s.PinSetting.AnalogChangeFrequency, s.PinSetting.Fn4,
			s.setting.Debounce, s.PinSetting.Reverse, s.PinSetting.Fn5,
			s.setting.DisableSqueeze, s.PinSetting.I2CSDA, s.PinSetting.Fn6,
			s.setting.DisableAutoSpace, s.PinSetting.Reset, s.PinSetting.I2CSCL,
		)
		str = strings.ReplaceAll(str, "\n", "\r\n")
		fmt.Printf("%v", str)
		*none = false
	}
}

func NewSetting() Setting {
	return Setting{
		ID:                  magicNumber,    // これは データの有効性を判定するためのマジックナンバー。flashメモリ上にあるデータが正しく初期化されているかの判定に使う。
		PinSetting:          NewPinAssign(), // GPIOのピン設定。
		Reverse:             false,          // 長短のパドル反転。初期値false。
		Speed:               20,             // 速度。初期値20WPM。
		Frequency:           700,            // 正弦波周波数。初期値440Hz。
		Debounce:            20,             // デバウンス期間。初期値20us。
		DashRate:            3.0,            // 長音比率。初期値3.0。
		EnableNegativeLogic: false,          // 負論理を使用するか。初期値false。
		EnableAnalogOutput:  false,          // アナログ出力による制限は出力を試みるか。初期値false。
		DisableSqueeze:      false,          // 長短パドルを同時に押したときのスクイーズ機能を無効化するか。初期値false。
		DisableAutoSpace:    false,          // 自動で無入力期間を入れる機能を無効化するか。初期値false。
	}
}

// GPIOのピンアサイン設定。使用しない場合-1に設定する。基本自由に変更してよい。アナログ入出力が必要な機能のみ26～29を使用すること。
type PinAssign struct {
	InputDit  int `json:"input-dit"`  // 短音入力ピン。
	InputDash int `json:"input-dash"` // 長音入力ピン。
	InputAny  int `json:"input-any"`  // 任意入力ピン。ストレートキー用ピン。
	DecodeCW  int `json:"decode-cw"`  // デコード用CW入力ピン。外部からのCWをでコードしてUSB Serialに文字として出力する。
	Reverse   int `json:"reverse"`    // 長短音ピン反転ピン。
	Reset     int `json:"reset"`      // 設定リセットピン。

	Fn1 int `json:"macro-1"` // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録1を指定する。 / (1秒以上) 設定値変更:スピード変更状態にする。
	Fn2 int `json:"macro-2"` // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録2を指定する。 / (1秒以上) 設定値変更:正弦波周波数変更状態にする。
	Fn3 int `json:"macro-3"` // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録3を指定する。 / (1秒以上) 設定値変更:デバウンス変更状態にする。
	Fn4 int `json:"macro-4"` // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録4を指定する。 / (1秒以上) 設定値変更:長音比率変更状態にする。
	Fn5 int `json:"macro-5"` // (1秒未満) 基本は何もしない。設定値変更状態で設定値を増加する。                / (1秒以上) 入力記録状態にする。長押ししながらFn1～4を短く押して記録先を指定する。その状態でFn5を離して打鍵すると記録される。再度Fn5を押して記録状態を終了する。1分以上記録終了されないとき記録を破棄して過去の記録状態を維持する。
	Fn6 int `json:"macro-6"` // (1秒未満) 基本は何もしない。設定値変更状態で設定値を減少する。                / (1秒以上) 記録出力状態にする。設定ファイルに記述がある文字列を出力する。長押ししながらFn1～4を短く押して定型文を出力する。

	ExternalInput1Dit  int `json:"external-input-1-dit"`  // 外部機器入力1短音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
	ExternalInput1Dash int `json:"external-input-1-dash"` // 外部機器入力1長音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
	ExternalInput1Any  int `json:"external-input-1-any"`  // 外部機器入力1任意ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
	ExternalInput2Dit  int `json:"external-input-2-dit"`  // 外部機器入力2短音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
	ExternalInput2Dash int `json:"external-input-2-dash"` // 外部機器入力2長音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
	ExternalInput2Any  int `json:"external-input-2-any"`  // 外部機器入力2任意ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
	ExternalInput3Dit  int `json:"external-input-3-dit"`  // 外部機器入力3短音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
	ExternalInput3Dash int `json:"external-input-3-dash"` // 外部機器入力3長音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
	ExternalInput3Any  int `json:"external-input-3-any"`  // 外部機器入力3任意ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。

	Output                int `json:"output"`                  // (矩形波)出力ピン。
	OutputSine            int `json:"output-sine"`             // モニター用正弦波出力ピン。アナログ出力する場合、26～29の何れかでなければならない。
	AnalogChangeSpeed     int `json:"analog-change-speed"`     // スピード変更。    可変抵抗等を使用して電圧値を変更して入力する。アナログ入力の為、26～29の何れかでなければならない。
	AnalogChangeFrequency int `json:"analog-change-frequency"` // 正弦波周波数変更。可変抵抗等を使用して電圧値を変更して入力する。アナログ入力の為、26～29の何れかでなければならない。
}

func NewPinAssign() PinAssign {
	// -1を指定した場合使用しない。
	return PinAssign{
		InputDit:  0, // 短音入力ピン。
		InputDash: 1, // 長音入力ピン。
		InputAny:  2, // 任意入力ピン。ストレートキー用ピン。
		DecodeCW:  3, // デコード用CW入力ピン。外部からのCWをでコードしてUSB Serialに文字として出力する。
		Reverse:   4, // 長短音ピン反転ピン。
		Reset:     5, // 設定リセットピン。

		SDA: 6, // I2C0
		SCL: 7, // I2C0

		// 8,9

		Fn1: 10, // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録1を指定する。 / (1秒以上) 設定値変更:スピード変更状態にする。
		Fn2: 11, // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録2を指定する。 / (1秒以上) 設定値変更:正弦波周波数変更状態にする。
		Fn3: 12, // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録3を指定する。 / (1秒以上) 設定値変更:デバウンス変更状態にする。
		Fn4: 13, // (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録4を指定する。 / (1秒以上) 設定値変更:長音比率変更状態にする。
		Fn5: 14, // (1秒未満) 基本は何もしない。設定値変更状態で設定値を増加する。                / (1秒以上) 入力記録状態にする。長押ししながらFn1～4を短く押して記録先を指定する。その状態でFn5を離して打鍵すると記録される。再度Fn5を押して記録状態を終了する。1分以上記録終了されないとき記録を破棄して過去の記録状態を維持する。
		Fn6: 15, // (1秒未満) 基本は何もしない。設定値変更状態で設定値を減少する。                / (1秒以上) 記録出力状態にする。設定ファイルに記述がある文字列を出力する。長押ししながらFn1～4を短く押して定型文を出力する。

		// 16,17,18,19,20,21,22

		// 23,24,25は露出していないことに注意。

		Output:                26, // (矩形波)出力ピン。
		OutputSine:            27, // モニター用正弦波出力ピン。PWM出力なので外部にLPHが必要。
		AnalogChangeSpeed:     28, // スピード変更。アナログ入出力ピン26, 27, 28, 29の何れかでなければならない。
		AnalogChangeFrequency: 29, // 正弦波周波数変更。アナログ入出力ピン26, 27, 28, 29の何れかでなければならない。
	}
}
