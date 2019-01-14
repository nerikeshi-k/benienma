package broker

import (
	"time"

	"github.com/nerikeshi-k/benienma/config"

	"github.com/gomodule/redigo/redis"
)

var pool *redis.Pool

func createPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				config.Get().Redis.Address,
				redis.DialPassword(config.Get().Redis.Password),
				redis.DialDatabase(config.Get().Redis.DB))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

const objectKeyPrefix string = "object:"
const objectCachedNameKeyPrefix string = "object-cached-name:" // キャッシュするときに付けたファイル名からの逆引き用
const lastRequestedTimeRankingKey string = "last-requested-time:ranking"

// GetObjectMetadata key に合致するオブジェクトをUnmarshalして返す
func GetObjectMetadata(key string) (Object, error) {
	var obj Object
	if pool == nil {
		pool = createPool()
	}
	conn := (*pool).Get()
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", objectKeyPrefix+key))
	if err != nil {
		return obj, err
	}
	obj.UnmarshalBinary(reply)
	return obj, nil
}

// GetObjectMetadataByCachedName キャッシュされてるファイル実体のファイル名から逆引きでObject Metadataを取る
func GetObjectMetadataByCachedName(cachedName string) (Object, error) {
	var obj Object
	if pool == nil {
		pool = createPool()
	}
	conn := (*pool).Get()
	defer conn.Close()
	objectIdentity, err := redis.String(conn.Do("GET", objectCachedNameKeyPrefix+cachedName))
	if err != nil {
		return obj, err
	}

	reply, err := redis.Bytes(conn.Do("GET", objectKeyPrefix+objectIdentity))
	if err != nil {
		return obj, err
	}

	obj.UnmarshalBinary(reply)
	return obj, nil

}

// SetObjectMetadata オブジェクトのメタデータをkeyに保存する
func SetObjectMetadata(key string, obj Object, expire time.Duration) error {
	if pool == nil {
		pool = createPool()
	}
	conn := (*pool).Get()
	defer conn.Close()
	encodedObj, err := obj.MarshalBinary()
	if err != nil {
		return err
	}
	err = conn.Send("MULTI")
	if err != nil {
		return err
	}

	// メイン、GCSのキーと同じ名前でObjectを登録
	err = conn.Send("SET", objectKeyPrefix+key, encodedObj)
	if err != nil {
		return err
	}
	err = conn.Send("EXPIRE", objectKeyPrefix+key, int(expire.Seconds()))
	if err != nil {
		return err
	}

	// キャッシュしたオブジェクトにつけたファイル名からの逆引き用
	err = conn.Send("SET", objectCachedNameKeyPrefix+obj.CachedName, key)
	if err != nil {
		return err
	}
	err = conn.Send("EXPIRE", objectCachedNameKeyPrefix+obj.CachedName, int(expire.Seconds()))
	if err != nil {
		return err
	}

	// 最終アクセス日時のランキング
	err = conn.Send("ZADD", lastRequestedTimeRankingKey, obj.LastRequestedAt.Unix(), obj.Identity)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil

}

// DeleteObjectMetadata オブジェクトのメタデータを消す
// キャッシュされたファイルの実体の削除はしない
func DeleteObjectMetadata(key string) error {
	if pool == nil {
		pool = createPool()
	}
	conn := (*pool).Get()
	defer conn.Close()

	var obj Object
	reply, err := redis.Bytes(conn.Do("GET", objectKeyPrefix+key))
	if err != nil {
		return err
	}
	obj.UnmarshalBinary(reply)

	err = conn.Send("MULTI")
	if err != nil {
		return err
	}

	err = conn.Send("DEL", objectKeyPrefix+key)
	if err != nil {
		return err
	}

	err = conn.Send("DEL", objectCachedNameKeyPrefix+obj.CachedName, key)
	if err != nil {
		return err
	}

	err = conn.Send("ZREM", lastRequestedTimeRankingKey, key)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

// DeleteOldObjects アクセスrankingを参照して古い順に消していく
func DeleteOldObjects(count int) error {
	if pool == nil {
		pool = createPool()
	}
	conn := (*pool).Get()
	defer conn.Close()

	for {
		identityList, err := redis.Strings(conn.Do("ZRANGE", lastRequestedTimeRankingKey, 0, 10))
		if err != nil {
			return err
		}
		var deleteCount int
		for _, identity := range identityList {
			err := DeleteObjectMetadata(identity)
			if err != nil {
				if err == redis.ErrNil {
					return err
				}
				err = conn.Send("ZREM", lastRequestedTimeRankingKey, identity)
				if err != nil {
					return err
				}
			} else {
				deleteCount++
			}
			if deleteCount > count {
				return nil
			}
		}
	}
}
