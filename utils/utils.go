package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// DoesFileExist ファイルの存在を調べる。あればtrue, なければfalse
func DoesFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// ListDir 与えられたパスのディレクトリのファイル一覧を返す
func ListDir(dir string) ([]string, error) {
	filenames := []string{}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return filenames, err
	}

	for _, file := range files {
		filenames = append(filenames, file.Name())
	}
	return filenames, nil
}

var beniDirPath *string

// GetBeniDirPath benienmaのプログラムがいるディレクトリの絶対パスを返す
func GetBeniDirPath() (string, error) {
	if beniDirPath != nil {
		return *beniDirPath, nil
	}
	exec, err := os.Executable()
	if err != nil {
		return "", err
	}

	path := filepath.Dir(exec)
	beniDirPath = &path
	return *beniDirPath, nil
}

// GenerateSafeName UUIDで作った文字列を返す
func GenerateSafeName() string {
	/**
	  bucket内でのobjectの名前にはスラッシュが入ったりしてるので使わない
	  （簡単のため一つのディレクトリにキャッシュを入れ子にせず保存したい）
	  UUIDだけで作ってるけど日付とかprefixでつけるのもありかも
	  **/
	uuid := uuid.New()
	return uuid.String()
}

// ExtractExtension 与えられたパスやファイル名から末尾の拡張子を取得して返す
func ExtractExtension(name string) string {
	/**
	  extract extension (ex. .png, .jpg) from filename
	  **/
	pos := strings.LastIndex(name, ".")
	return (name[pos:])
}
