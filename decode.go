package main

import (
	"errors"
)

var (

	// モールス信号と英数字記号の対応表
	// .が0、-が1としている。
	morseCodeMap = map[string]string{
		"01":      "A",
		"1000":    "B",
		"1010":    "C",
		"100":     "D",
		"0":       "E",
		"0010":    "F",
		"110":     "G",
		"0000":    "H",
		"00":      "I",
		"0111":    "J",
		"101":     "K",
		"0100":    "L",
		"11":      "M",
		"10":      "N",
		"111":     "O",
		"0110":    "P",
		"1101":    "Q",
		"010":     "R",
		"000":     "S",
		"1":       "T",
		"001":     "U",
		"0001":    "V",
		"011":     "W",
		"1001":    "X",
		"1011":    "Y",
		"1100":    "Z",
		"11111":   "0",
		"01111":   "1",
		"00111":   "2",
		"00011":   "3",
		"00001":   "4",
		"00000":   "5",
		"10000":   "6",
		"11000":   "7",
		"11100":   "8",
		"11110":   "9",
		"010101":  ".",
		"110011":  ",",
		"001100":  "?",
		"101011":  "!",
		"10110":   "(",
		"101101":  ")",
		"011110":  "'",
		"100001":  "-",
		"10010":   "/",
		"010010":  "\"",
		"011010":  "@",
		"0001001": "$",
		"01000":   "&",

		"dit":      "dit",      // 来ない。
		"dash":     "dash",     // 来ない。
		"straight": "straight", // 来ない。
		"*":        "straight", // ストレートキーは長いのでこれにしている。
		"reset":    "reset",
		"sp_up":    "speed up",
		"sp_dn":    "speed down",
		"fq_up":    "frequency up",
		"fq_dn":    "frequency down",
		"de_up":    "debounce up",
		"de_dn":    "debounce down",
		"reverse":  "reverse",
	}
)

func ReadBuf(buf *[]ePushState) string {
	// 短音を0,長音を1として文字列にする。
	morseCode := ""
	for _, b := range *buf {
		if b == PUSH_DIT {
			morseCode += "0"
		} else if b == PUSH_DASH {
			morseCode += "1"
		} else if b == PUSH_STRAIGHT {
			// ストレートキー。
			morseCode += "*"
		} else {
			// 設定変更ピンなど。
			morseCode += b.String()
		}
	}
	char, err := MorseToChar(morseCode)
	if err != nil {
		//fmt.Printf("Error: %v\n", err)
		char = "error"
	}
	*buf = (*buf)[:0] // ターミナル表示用バッファをクリア。capacityは維持。
	//log.Printf("buf: %#v,\tmorseCode: %#v,\tchar: %#v", *buf, morseCode, char)
	return char
}

// MorseToChar はモールス信号を英数字または記号に変換する関数
func MorseToChar(morseCode string) (string, error) {
	// 対応するモールス信号が存在するか確認
	if char, ok := morseCodeMap[morseCode]; ok {
		return char, nil
	}
	// マッピングが存在しない場合はエラーを返す
	return "", errors.New("invalid morse code")
}
