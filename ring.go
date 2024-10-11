package main

import (
	"container/ring"
)

type InputType int // 受理した入力の種類。

const (
	Push_No InputType = iota
	Push_InputDit
	Push_InputDash
	Push_InputAny
	Push_DecodeCW
	Push_Reverse
	Push_Reset
	Push_Fn1
	Push_Fn2
	Push_Fn3
	Push_Fn4
	Push_Fn5
	Push_Fn6
	Change_AnalogChangeSpeed
	Change_AnalogChangeFrequency
)

type RingBuffer struct {
	r        *ring.Ring // 受理した入力の履歴。
	capacity int        // リングバッファのサイズ
	count    int        // 追加された要素数
}

// 使用例
// func main() {
// 	// サイズ5のリングバッファを作成
// 	rb := NewInputRingBuffer(5)
//
// 	// InputTypeの値をリングバッファに追加
// 	rb.Add(Push_InputDit)
// 	rb.Add(Push_InputDash)
// 	rb.Add(Push_InputAny)
//
// 	// 順方向での要素表示
// 	rb.Do(func(value InputType) {
// 		fmt.Println(value)
// 	})
//
// 	// 要素数を確認
// 	fmt.Printf("要素数: %d\n", rb.Count())
//
// 	// リングバッファをリセット
// 	rb.Reset()
//
// 	// リセット後の要素数を確認
// 	fmt.Printf("要素数: %d\n", rb.Count())
// }

// NewInputRingBufferは指定されたサイズのリングバッファを初期化します
func NewInputRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		r:        ring.New(size),
		capacity: size,
		count:    0,
	}
}

// AddはリングバッファにInputTypeの値を追加します
func (rb *RingBuffer) Add(value InputType) {
	rb.r.Value = value
	rb.r = rb.r.Next()
	if rb.count < rb.capacity {
		rb.count++
	}
}

// Countはリングバッファ内の要素数を返します
func (rb *RingBuffer) Count() int {
	return rb.count
}

// Resetはリングバッファ内の要素数を0にリセットします。値はクリアしません。
func (rb *RingBuffer) Reset() {
	rb.count = 0
}

// Doはリングバッファ内の要素数だけ順方向に関数を実行します
// 順方向での要素表示する例。
//
//	rb.Do(func(value InputType) {
//		fmt.Println(value)
//	})
func (rb *RingBuffer) Do(f func(InputType)) {
	current := rb.r
	for i := 0; i < rb.count; i++ {
		if current.Value != nil {
			f(current.Value.(InputType))
		}
		current = current.Next()
	}
}

// DoReverseはリングバッファ内の要素数だけ逆方向に関数を実行します
// 逆方向での要素表示する例
//
//	rb.DoReverse(func(value InputType) {
//		fmt.Println(value)
//	})
func (rb *RingBuffer) DoReverse(f func(InputType)) {
	current := rb.r.Prev()
	for i := 0; i < rb.count; i++ {
		if current.Value != nil {
			f(current.Value.(InputType))
		}
		current = current.Prev()
	}
}
