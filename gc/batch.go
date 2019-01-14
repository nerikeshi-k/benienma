package gc

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/nerikeshi-k/benienma/broker"
	"github.com/nerikeshi-k/benienma/config"
	"github.com/nerikeshi-k/benienma/utils"
)

var locked bool

func init() {
	locked = false
}

func StartCollecting() {
	for {
		time.Sleep(time.Duration(config.Get().Collect.Span) * time.Second)
		if locked {
			continue
		}
		(func() {
			locked = true
			defer (func() { locked = false })()

			err := checkVolumeAndSweep()
			if err != nil {
				fmt.Printf("SweepError: %v", err)
			}

			err = collectUnreferencedObjects()
			if err != nil {
				fmt.Printf("CollectError: %v", err)
			}
		})()
	}
}

func collectUnreferencedObjects() error {
	/**
	  Redisレコードから参照されなくなったキャッシュファイルを消去する
	  **/

	// CacheDirに存在しているオブジェクト取得
	cachedObjectNames, err := utils.ListDir(config.Get().CacheDirPath)
	if err != nil {
		return err
	}

	// CacheDirに実体はあるがredisのレコードから参照されなくなっているものを見つけたら
	// 削除
	for _, cachedName := range cachedObjectNames {
		_, err = broker.GetObjectMetadataByCachedName(cachedName)
		if err == redis.ErrNil {
			if err := os.Remove(path.Join(config.Get().CacheDirPath, cachedName)); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkVolumeAndSweepで使うsortで必要なstruct
type objKeyByLastReqTime struct {
	Key             string
	FileName        string
	LastRequestedAt time.Time
}
type objKeysByLastReqTime []objKeyByLastReqTime

func (o objKeysByLastReqTime) Len() int {
	return len(o)
}
func (o objKeysByLastReqTime) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}
func (o objKeysByLastReqTime) Less(i, j int) bool {
	return o[i].LastRequestedAt.Before(o[j].LastRequestedAt)
}

func checkVolumeAndSweep() error {
	/**
	  保存先ディレクトリの容量がmax_volumeを超えそうであれば、
	  最終アクセスの遅い順にファイルを削除し、状態を解消させる。
	**/

	for {
		cacheVolume := dirSizeMB(config.Get().CacheDirPath)
		// if volume is over 90% of max
		if cacheVolume > (float64(config.Get().MaxVolume) * 0.9) {
			err := broker.DeleteOldObjects(1)
			if err != nil {
				return err
			}
			collectUnreferencedObjects()
		} else {
			break
		}
	}

	return nil
}

func dirSizeMB(path string) float64 {
	// from https://stackoverflow.com/questions/32482673/golang-how-to-get-directory-total-size
	var dirSize int64

	readSize := func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			dirSize += file.Size()
		}
		return nil
	}

	filepath.Walk(path, readSize)
	sizeMB := float64(dirSize) / 1024.0 / 1024.0

	return sizeMB
}
