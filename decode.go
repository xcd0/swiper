package main

import (
	"fmt"
	"strings"
)

// モールス信号の種類
type WordType int

const (
	TypeMorse WordType = iota
	TypeChangeSetting
	TypeOther
)

// モールスモードの定義
type MorseMode int

const (
	ModeReadMorse   MorseMode = iota // 英文モールス
	ModeReadJPMorse                  // 和文モールス
)

type MorseMap map[MorseMode]string

// モールス信号デコード情報
type WordInfo struct {
	Type   WordType
	String MorseMap
}

// デコード用のマップ
type DecodeMap map[string]WordInfo

// デコードマップの定義（英文モールスと和文モールス）
var decode_map = DecodeMap{

	".-":   {Type: TypeMorse, String: MorseMap{ModeReadMorse: "A", ModeReadJPMorse: "イ"}},
	"-...": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "B", ModeReadJPMorse: "ハ"}},
	"-.-.": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "C", ModeReadJPMorse: "ニ"}},
	"-..":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "D", ModeReadJPMorse: "ホ"}},
	".":    {Type: TypeMorse, String: MorseMap{ModeReadMorse: "E", ModeReadJPMorse: "ヘ"}},
	"..-.": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "F", ModeReadJPMorse: "チ"}},
	"--.":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "G", ModeReadJPMorse: "リ"}},
	"....": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "H", ModeReadJPMorse: "ヌ"}},
	"..":   {Type: TypeMorse, String: MorseMap{ModeReadMorse: "I", ModeReadJPMorse: "゛"}},
	".---": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "J", ModeReadJPMorse: "ヲ"}},
	"-.-":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "K", ModeReadJPMorse: "ワ"}}, // 送信要求の意味でも使われる。
	".-..": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "L", ModeReadJPMorse: "カ"}},
	"--":   {Type: TypeMorse, String: MorseMap{ModeReadMorse: "M", ModeReadJPMorse: "ヨ"}},
	"-.":   {Type: TypeMorse, String: MorseMap{ModeReadMorse: "N", ModeReadJPMorse: "タ"}},
	"---":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "O", ModeReadJPMorse: "レ"}},
	".--.": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "P", ModeReadJPMorse: "ツ"}},
	"--.-": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "Q", ModeReadJPMorse: "ネ"}},
	".-.":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "R", ModeReadJPMorse: "ナ"}},
	"...":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "S", ModeReadJPMorse: "ラ"}},
	"-":    {Type: TypeMorse, String: MorseMap{ModeReadMorse: "T", ModeReadJPMorse: "ム"}},
	"..-":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "U", ModeReadJPMorse: "ウ"}},
	"...-": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "V", ModeReadJPMorse: "ク"}},
	".--":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "W", ModeReadJPMorse: "ヤ"}},
	"-..-": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "X", ModeReadJPMorse: "マ"}}, // 乗算の意味でも使われる。
	"-.--": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "Y", ModeReadJPMorse: "ケ"}},
	"--..": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "Z", ModeReadJPMorse: "フ"}},

	// 数字
	"-----": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "0"}},
	".----": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "1"}},
	"..---": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "2"}},
	"...--": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "3"}},
	"....-": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "4"}},
	".....": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "5"}},
	"-....": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "6"}},
	"--...": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "7"}},
	"---..": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "8"}},
	"----.": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "9"}},

	// 記号
	".-.-.-": {Type: TypeMorse, String: MorseMap{ModeReadMorse: ".", ModeReadJPMorse: "、"}},  // ピリオド
	"--..--": {Type: TypeMorse, String: MorseMap{ModeReadMorse: ","}},                        // カンマ
	"..--..": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "?"}},                        // クエスチョンマーク
	".----.": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "'"}},                        // アポストロフィ
	"-.-.--": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "!"}},                        // 感嘆符
	"-..-.":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "/", ModeReadJPMorse: "モ"}},  // スラッシュ
	"-.--.":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "(", ModeReadJPMorse: "ル"}},  // 開き括弧
	"-.--.-": {Type: TypeMorse, String: MorseMap{ModeReadMorse: ")", ModeReadJPMorse: "（"}},  // 閉じ括弧 // 和文モールス中にアルファベットを含めるときは前後を（）で括る。
	".-...":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "&", ModeReadJPMorse: "オ"}},  // アンパサンド 待機要求の意味でも使われる。
	"---...": {Type: TypeMorse, String: MorseMap{ModeReadMorse: ":"}},                        // コロン
	"-.-.-.": {Type: TypeMorse, String: MorseMap{ModeReadMorse: ";"}},                        // セミコロン
	"-...-":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "=", ModeReadJPMorse: "メ"}},  // イコール 送信開始の意味でも使われる。
	".-.-.":  {Type: TypeMorse, String: MorseMap{ModeReadMorse: "+", ModeReadJPMorse: "ン"}},  // プラス 送信終了の意味でも使われる。
	"-....-": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "-"}},                        // マイナス
	"..--.-": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "_"}},                        // アンダースコア
	".-..-.": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "\"", ModeReadJPMorse: "）"}}, // ダブルクオート
	".--.-.": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "@"}},                        // アットマーク

	"......":   {Type: TypeMorse, String: MorseMap{ModeReadMorse: "^"}},                              // べき乗"^"
	"........": {Type: TypeMorse, String: MorseMap{ModeReadMorse: "訂正"}},                             // 訂正 ※「HH」と表現される
	"...-.-":   {Type: TypeMorse, String: MorseMap{ModeReadMorse: "通信終了"}},                           // 通信の終了 ※「VA」と表現される
	"...-.":    {Type: TypeMorse, String: MorseMap{ModeReadMorse: "了解", ModeReadJPMorse: "[訂正・終了]"}}, // ラタ

	// 和文残り
	".-.-":   {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ロ"}},
	"..-..":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ト"}},
	"---.":   {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ソ"}},
	".-..-":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ヰ"}},
	"..--":   {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ノ"}},
	"----":   {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "コ"}},
	"-.---":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "エ"}},
	".-.--":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "テ"}},
	"--.--":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ア"}},
	"-.-.-":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "サ"}},
	"-.-..":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "キ"}},
	"-..--":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ユ"}},
	"..-.-":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ミ"}},
	"--.-.":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "シ"}},
	".--..":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ヱ"}},
	"--..-":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ヒ"}},
	".---.":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "セ"}},
	"---.-":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ス"}},
	"..--.":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "゜"}},
	".--.-":  {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "ー"}},
	".-.-..": {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "」"}},
	"-..---": {Type: TypeMorse, String: MorseMap{ModeReadJPMorse: "[本文]"}}, // ホレ
}

func DecodeMorse(rb *InputRingBuffer, mode MorseMode) (string, error) {
	if rb.Count() == 0 {
		return "", fmt.Errorf("リングバッファが空です。")
	}
	var morseCode strings.Builder
	rb.Do(func(input InputType) {
		switch input {
		case Push_InputDit:
			morseCode.WriteString(".")
		case Push_InputDash:
			morseCode.WriteString("-")
		}
	})
	code := morseCode.String()
	if info, ok := decode_map[code]; ok {
		if info.Type == TypeMorse {
			if str, ok := info.String[mode]; ok {
				return str, nil
			}
			// モードに対応する文字がない場合、別のモードの文字を使用
			for _, v := range info.String {
				if len(v) != 0 {
					return v, nil
				}
			}
		}
	}
	return "", fmt.Errorf("該当する符号が見つかりませんでした。:%v", code) // 該当なし
}
