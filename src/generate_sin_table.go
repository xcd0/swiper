//go:build ignore

// このファイルはgo generateで使用される。
// 実行コマンド: go generate

package main

import (
	"fmt"
	"log"
	"math"
	"os"
)

const (
	steps = 1000 // 正弦波のステップ数
)

func main() {
	// 出力ファイルを開く
	file, err := os.Create("sin_table.go")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if false {
		// Goコードとしてテーブルを書き出す
		fmt.Fprintln(file, "package main")
		fmt.Fprintln(file)
		fmt.Fprintf(file, "// 正弦波のルックアップテーブル (%dステップ)\n", steps)
		fmt.Fprintln(file, "var sinTable = []uint32{")

		// 正弦波の値を計算しテーブルとして書き込む
		for i := 0; i < steps; i++ {
			// math.MaxUint32を使用して、0からMaxUint32の範囲にスケーリング

			den := (math.Sin(2.*math.Pi*float64(i)/float64(1000)) + 1.) / 2. // -1~1を0~1に変換
			div := float64(math.MaxUint32)
			value := uint32(den * div) // 0～MaxUint32にマッピング

			// 確認用ログ出力
			if false {
				if i%50 == 0 || i == 999 {
					log.Printf("i:%4d : %.3f * %.0f = %12.1f (%3.1f %%)\n", i, den, div, value, float64(value)/float64(math.MaxUint32)*100)
				}
			}

			fmt.Fprintf(file, "\t%d,\n", value)
		}
		fmt.Fprintln(file, "}")
	} else {
		// Goコードとしてテーブルを書き出す
		fmt.Fprintln(file, "package main")
		fmt.Fprintln(file)
		fmt.Fprintf(file, "// 正弦波のルックアップテーブル (%dステップ)\n", steps)
		fmt.Fprintf(file, "// 割合で定義する。\n")
		fmt.Fprintln(file, "var sinTable = []float32{")

		// 正弦波の値を計算しテーブルとして書き込む
		for i := 0; i < steps; i++ {
			// math.MaxUint32を使用して、0からMaxUint32の範囲にスケーリング

			den := (math.Sin(2.*math.Pi*float64(i)/float64(1000)) + 1.) / 2. // -1~1を0~1に変換
			value := float32(den)

			fmt.Fprintf(file, "\t%v,\n", value)
		}
		fmt.Fprintln(file, "}")
	}
}
