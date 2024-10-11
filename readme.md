# raspberry pi picoで作る複式電鍵用エレキー

複式電鍵で練習したかった。  
電鍵もエレキーも高いので自作する。  
実用に耐えるかどうかは...  

## 概要

[keyer](https://en.wikipedia.org/wiki/Keyer)やエレキーと呼ばれるモールス信号入力補助装置。

## 参考資料

Arduino用のOSSがある。

https://github.com/k3ng/k3ng_cw_keyer/wiki

上記を参考に実装する機能を決めていく。

## 機能

- 入力
	- ストレートキー
	- 短音パドル
	- 長音パドル
- 出力
	- 矩形波
	- 正弦波


## ピンアサイン <!--{{{-->


初期状態のピンアサイン。任意に変更可能。
但し、raspberry pi picoの制約上、23,24,25は使用不可、アナログ入出力は26,27,28,29のみ可能なことに注意。  

 GPIO  | 説明
 ----- | ------------------------------
   0   | 短音入力ピン。
   1   | 長音入力ピン。
   2   | 任意入力ピン。ストレートキー用ピン。
   3   | 設定リセットピン。
   4   | 長短音ピン反転ピン。
   5   | (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録1を指定する。 <br> (1秒以上) 設定値変更:スピード変更状態にする。
   6   | (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録2を指定する。 <br> (1秒以上) 設定値変更:正弦波周波数変更状態にする。
   7   | (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録3を指定する。 <br> (1秒以上) 設定値変更:デバウンス変更状態にする。
   8   | (1秒未満) 基本は何もしない。入力記録状態または記録出力状態で記録4を指定する。 <br> (1秒以上) 設定値変更:長音比率変更状態にする。
   9   | (1秒未満) 基本は何もしない。設定値変更状態で設定値を増加する。                <br> (1秒以上) 入力記録状態にする。長押ししながらMacro1～4を短く押して記録先を指定する。その状態でMacro5を離して打鍵すると記録される。再度Macro5を押して記録状態を終了する。1分以上記録終了されないとき記録を破棄して過去の記録状態を維持する。
  10   | (1秒未満) 基本は何もしない。設定値変更状態で設定値を減少する。                <br> (1秒以上) 記録出力状態にする。設定ファイルに記述がある文字列を出力する。長押ししながらMacro1～4を短く押して定型文を出力する。
  11   | 未使用。
  12   | 未使用。
  13   | デコード用CW入力ピン。外部からのCWをでコードしてUSB Serialに文字として出力する。
  14   | 外部機器入力1短音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
  15   | 外部機器入力1長音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
  16   | 外部機器入力1任意ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
  17   | 外部機器入力2短音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
  18   | 外部機器入力2長音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
  19   | 外部機器入力2任意ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
  20   | 外部機器入力3短音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
  21   | 外部機器入力3長音ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
  22   | 外部機器入力3任意ピン。初期状態では使用しない。別の電鍵からの入力を読み込んでkeyerとして動作させたいとき使用する。
  23   | 使用不可。露出していない。
  24   | 使用不可。露出していない。
  25   | 使用不可。露出していない。
  26   | (矩形波)出力ピン。
  27   | モニター用正弦波出力ピン。
  28   | スピード変更。アナログ入力なので26, 27, 28, 29の何れかでなければならない。
  29   | 正弦波周波数変更。アナログ入力なので26, 27, 28, 29の何れかでなければならない。

> ![](./img/pico.png)

<!--}}}-->

## 仕様

### 信号の速さ

初期値を25wpmとして、調整可能範囲を5wpmから35wpmまでとしている。  

### モニター用ビープ音

とりあえず、800を基準に、400から1200まで50Hz刻みで設定できるようにしている。  

一応PWM変調で正弦波を出力している。
周波数はそこそこずれる。800の時実測850くらいだった。
この辺はPWM変調の使い方が悪そう。
ただ、開始終了部分に窓関数をかけていないので音のなり始めと終わりがひどい。
なんでもいいので窓関数をかける必要がある。


#### 正弦波出力の改善
- [R2R DAC](https://en.wikipedia.org/wiki/Resistor_ladder)を外部に作る？ 
	- 8bitならGPIOを8ポート占有することになる。
	- [参考](https://www.instructables.com/Arbitrary-Wave-Generator-With-the-Raspberry-Pi-Pic/)
- [秋月で売っているDAC](https://akizukidenshi.com/catalog/c/cdaconver/)
	- [MCP4726](https://akizukidenshi.com/catalog/g/g107995/)は在庫があったはず。
	- I2Cでつなぐ必要がある。 [参考](https://webmidiaudio.com/npage313.html)
	- 他にLCDなどを実装する可能性を考えるとI2Cは悪くない選択肢。
	- I2Cは 2ポートあればよい。(SDA,SCL,Vdd,GND)
		- 但しペアがある。 https://tinygo.org/docs/reference/microcontrollers/machine/pico/#type-i2c
		I2C | GPIO (SDA, SCL)
		----|------
		I2C0|(0,1),(4,5),(8,9),(12,13),(16,17),(20,21)
		I2C1|(2,3),(6,7),(10,11),(14,15),(18,19),(26,27)


### チャタリング防止

デバウンス初期値20msとしている。
0msから200msまで10ms刻みで設定できる。

### 長短パドル反転

反転できるようにしている。

### 長押し

長押ししているとリピートする。

### 正弦波出力

モニター用に正弦波を出している。
と言っても、PWM変調を利用した出力でしかないので、外部にLPFを必要とする。  
LPFがなくても単純な[圧電サウンダ](https://akizukidenshi.com/catalog/g/g101251/)では単純につなげばそれっぽく音が鳴る。  
勿論音は良くない。  

例えば、[ブレッドボード用ダイナミックスピーカー](https://akizukidenshi.com/catalog/g/g112587/) を使用して、  
1mの距離で40dB程度で鳴らしたい場合、ChatGPTによると28mV程度でよい模様。合っているかどうか不明。  
LPFの出力が3V程度だとするとそのまま使用すると音が大きすぎると思われる。
なので可変抵抗を使用して調節できるようにするとそう。  
[スイッチ付 小型ボリューム 10kΩB](https://akizukidenshi.com/catalog/g/g117281/) を使う場合、100kΩ程度の抵抗R1を使って分圧すれば良さそう。  
![](./img/bad_speaker_voltage_divider.png)  
しかしよく考えるとこれだとスピーカーに直流電流が流れてしまうのでよくなさそう。  

https://github.com/martinkooij/pi-pico-tone に

> ![](./img/tone-line-output.png)  

のように、LPFは16kHz、HPFが2Hzくらいで良さげな回路があった。  
このスピーカーを動かすだけなら出力部分のインピーダンスマッチング用抵抗は不要と思われる。  

### 設定保存

設定値保存機能を実装した。  
flashにあまり頻繁に書き込まないように、設定値変更後最初の打鍵時に保存するようにしている。


## おまけ機能

Tera Term等でシリアル覗くと情報が見える。

1. 打鍵した短音と長音を解析して文字に変換して表示する。

![](./img/teraterm.png)

2. 5秒以上何も操作しないと現在の設定を出力する。

![](./img/teraterm_setting.png)

## 検討事項

### ディスプレイ

下記を買ってみたので届いたら試す。  
https://www.amazon.co.jp/dp/B0CP46CSWV/  
https://www.microfan.jp/2023/01/oled-display/  

tinygoにドライバがある。  
https://pkg.go.dev/tinygo.org/x/drivers/sh1106  


### 同時押しはどうするか。
	- 現状無視。
	- スクイーズ操作を実装する？

### ロータリーエンコーダ対応

設定値変更はロータリーエンコーダか単純な可変抵抗でやりたい。ボタンポチポチは面倒。  
ただし設定値保存ができる場合、そんなにはいじらないので十分ではある。  

