package main

import (
	"encoding/json"
	"fmt"
	"log"
	"machine"
	"os"
	"time"

	"tinygo.org/x/tinyfs/littlefs"
)

var (
	blockDevice      = machine.Flash
	filesystem       = littlefs.New(blockDevice)
	setting_filepath = "/settings.json"
)

// エラー時の無限ループ処理
func handleError(err error) {
	for {
		fmt.Printf("%v\r\n", err)
		time.Sleep(time.Second)
	}
}

// 設定をファイルから読み込む関数
func readSettingFromFile(fs *littlefs.LFS, setting_filepath string) (Setting, error) {
	log.Printf("readSettingFromFile")
	// ファイルをオープン
	f, err := fs.OpenFile(setting_filepath, os.O_RDONLY)
	if err != nil {
		return Setting{}, fmt.Errorf("readSettingFromFile: could not open file: %w", err)
	}
	defer f.Close()

	// ファイルからデータを読み込む
	var setting Setting
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&setting)
	if err != nil {
		return Setting{}, fmt.Errorf("failed to decode JSON: %w", err)
	}
	return setting, nil
}

// 設定をファイルに書き込む関数
func writeSettingToFile(fs *littlefs.LFS, setting Setting, setting_filepath string) error {
	// ファイルをオープン
	f, err := fs.OpenFile(setting_filepath, os.O_CREATE|os.O_WRONLY)
	if err != nil {
		return fmt.Errorf("writeSettingToFile: could not open file: %w", err)
	}
	defer f.Close()
	// シリアライズしてJSON形式で書き込む
	encoder := json.NewEncoder(f)
	err = encoder.Encode(setting)
	if err != nil {
		return fmt.Errorf("writeSettingToFile: could not encode JSON: %w", err)
	}
	return nil
}

func init_flash() {
	log.Printf("init_flash")
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
