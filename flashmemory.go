package main

import (
	"encoding/binary"
	"fmt"
	"machine"
	"os"
	"time"

	"tinygo.org/x/tinyfs/littlefs"
)

// マジックナンバーの定義
const magicNumber uint32 = 0xDEADBEEF

var (
	blockDevice      = machine.Flash
	filesystem       = littlefs.New(blockDevice)
	setting_filepath = "/settings.bin"
)

// エラー時の無限ループ処理
func handleError(err error) {
	for {
		fmt.Printf("%v\r\n", err)
		time.Sleep(time.Second)
	}
}

// シリアライズ関数
func serializeSetting(setting Setting) []byte {
	buf := make([]byte, 17) // マジックナンバー(4バイト) + SpeedOffset(4バイト) + FreqOffset(4バイト) + DebounceOffset(4バイト) + Reverse(1バイト)
	binary.LittleEndian.PutUint32(buf[0:], setting.ID)
	binary.LittleEndian.PutUint32(buf[4:], uint32(setting.SpeedOffset))
	binary.LittleEndian.PutUint32(buf[8:], uint32(setting.FreqOffset))
	binary.LittleEndian.PutUint32(buf[12:], uint32(setting.DebounceOffset))
	if setting.Reverse {
		buf[16] = 1
	} else {
		buf[16] = 0
	}
	return buf
}

// デシリアライズ関数
func deserializeSetting(buf []byte) Setting {
	setting := Setting{}
	setting.ID = binary.LittleEndian.Uint32(buf[0:])
	setting.SpeedOffset = int(binary.LittleEndian.Uint32(buf[4:]))
	setting.FreqOffset = int(binary.LittleEndian.Uint32(buf[8:]))
	setting.DebounceOffset = int(binary.LittleEndian.Uint32(buf[12:]))
	setting.Reverse = buf[16] == 1
	return setting
}

// 設定をファイルから読み込む関数
func readSettingFromFile(fs *littlefs.LFS, setting_filepath string) (Setting, bool, error) {
	// ファイルをオープン
	f, err := fs.OpenFile(setting_filepath, os.O_RDONLY)
	if err != nil {
		return Setting{}, false, fmt.Errorf("readSettingFromFile: could not open file: %w", err)
	}
	defer f.Close()

	// ファイルからデータを読み込む
	buf := make([]byte, 17)
	_, err = f.Read(buf)
	if err != nil {
		return Setting{}, false, fmt.Errorf("readSettingFromFile: failed to read file: %w", err)
	}

	// デシリアライズしてSettingに変換
	setting := deserializeSetting(buf)
	if setting.ID != magicNumber {
		return setting, false, nil // マジックナンバーが一致しない場合は無効
	}

	return setting, true, nil
}

// 設定をファイルに書き込む関数
func writeSettingToFile(fs *littlefs.LFS, setting Setting, setting_filepath string) error {
	// ファイルをオープン
	f, err := fs.OpenFile(setting_filepath, os.O_CREATE|os.O_WRONLY)
	if err != nil {
		return fmt.Errorf("writeSettingToFile: could not open file: %w", err)
	}
	defer f.Close()

	// シリアライズしたデータを書き込む
	data := serializeSetting(setting)
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("writeSettingToFile: could not write to file: %w", err)
	}

	return nil
}

func init_flash() {
	// littlefsの設定
	filesystem.Configure(&littlefs.Config{
		CacheSize:     512,
		LookaheadSize: 512,
		BlockCycles:   100,
	})

	// ファイルシステムのマウント
	if err := filesystem.Mount(); err != nil {
		fmt.Println("Filesystem not found, formatting...")
		// フォーマットしてから再マウント
		if err := filesystem.Format(); err != nil {
			handleError(fmt.Errorf("Could not format filesystem: %w", err))
		}
		if err := filesystem.Mount(); err != nil {
			handleError(fmt.Errorf("Could not mount filesystem after formatting: %w", err))
		}
	}
}
