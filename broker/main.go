package broker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/nerikeshi-k/benienma/config"
	"github.com/nerikeshi-k/benienma/utils"

	"cloud.google.com/go/storage"
)

// Object redisに保存されるメタデータ
type Object struct {
	Identity        string    `json:"Identity"`    // GCS の bucket 内での名前
	CachedName      string    `json:"cached_name"` // ローカルにキャッシュする際に付けられたファイル名
	Path            string    `json:"path"`        // キャッシュされたオブジェクトのパス
	Size            int64     `json:"size"`
	LastRequestedAt time.Time `json:"last_requested_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// MarshalBinary broker.Objectをredisに保存する際に必要
func (o *Object) MarshalBinary() ([]byte, error) {
	b, err := json.Marshal(o)
	return b, err
}

// UnmarshalBinary broker.Objectをredisに保存する際に必要
func (o *Object) UnmarshalBinary(data []byte) error {
	err := json.Unmarshal(data, &o)
	return err
}

// NotFound Objectが存在しなかったとき返すエラー
type NotFound struct {
	identity string
}

func (n *NotFound) Error() string {
	return fmt.Sprintf("%s is not found", n.identity)
}

// Get identityを手がかりにキャッシュから探し、なければGCSから探してキャッシュして返す
func Get(identity string) (*Object, error) {
	var object Object
	object, err := GetObjectMetadata(identity)
	if err != nil {
		if err != redis.ErrNil {
			return nil, err
		}
	} else {
		// キャッシュファイルが存在しなかった場合、Objectを削除して再読込
		if !utils.DoesFileExist(object.Path) {
			err := DeleteObjectMetadata(identity)
			if err != nil {
				return nil, err
			}
			return Get(identity)
		}
		// setし直してexpired dateを更新, メタデータを返す
		object.LastRequestedAt = time.Now()
		if err := SetObjectMetadata(identity, object, time.Hour*time.Duration(config.Get().ExpiredHours)); err != nil {
			return nil, err
		}
		return &object, nil
	}

	object, err = fetchObject(identity)
	if err == storage.ErrObjectNotExist {
		return nil, &NotFound{identity: identity}
	} else if err != nil {
		return &object, err
	}

	if err = SetObjectMetadata(identity, object, time.Hour*time.Duration(config.Get().ExpiredHours)); err != nil {
		return nil, err
	}
	return &object, nil
}
