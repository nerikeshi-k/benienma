package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/nerikeshi-k/benienma/utils"
)

var config Config

type Config struct {
	Port         int    `json:"port"`
	CacheDirPath string `json:"cached_objects_directory"`
	ExpiredHours int    `json:"expired_hours"`
	MaxVolume    int    `json:"max_volume"`
	Bucket       struct {
		Name string `json:"name"`
	} `json:"bucket"`
	Redis struct {
		Address  string `json:"address"`
		Password string `json:"password"`
		DB       int    `json:"db_id"`
	} `json:"redis"`
	Collect struct {
		Span int `json:"span_sec"`
	} `json:"collect"`
	LastModified string `json:"last_modified"`
}

// init 設定ファイルの読み込みをする
func init() {
	err := Load()
	if err != nil {
		panic(err)
	}
}

// Load コマンドライン引数を取って設定ファイルを読み込んだりする
func Load() error {
	var configJSONPath string
	beniDir, err := utils.GetBeniDirPath()
	if err != nil {
		return err
	}
	// 引数に指定されてないならbenienmaが動作しているディレクトリから探す
	flag.StringVar(&configJSONPath, "conf", path.Join(beniDir, "config.json"), "path of config.json")
	flag.Parse()
	if !utils.DoesFileExist(configJSONPath) {
		return fmt.Errorf("%s does not exist", configJSONPath)
	}
	bytes, err := ioutil.ReadFile(configJSONPath)
	if err != nil {
		return err
	}
	// JSONパース
	if err := json.Unmarshal(bytes, &config); err != nil {
		return err
	}
	// CacheDirの存在チェック
	if !utils.DoesFileExist(config.CacheDirPath) {
		return fmt.Errorf("Object caching dir %s does not exist", config.CacheDirPath)
	}
	return nil
}

// Get 設定構造体を返却する
func Get() Config {
	return config
}
